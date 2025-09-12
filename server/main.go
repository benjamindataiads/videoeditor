package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAddr = ":8080"
)

type Asset struct {
	ID          string  `json:"id"`
	Filename    string  `json:"filename"`
	URL         string  `json:"url"`
	Kind        string  `json:"kind"` // image | video | audio
	DurationSec float64 `json:"durationSec"`
}

type ExportClip struct {
	AssetID string `json:"assetId"`
	// Optional trims for videos (in seconds). Ignored for images.
	StartSec float64 `json:"startSec"`
	EndSec   float64 `json:"endSec"`
	// Duration for images in seconds
	DurationSec float64 `json:"durationSec"`
	// Whether to reverse/mirror the clip horizontally
	Reversed bool `json:"reversed"`
	// Whether to play the clip in reverse (backwards in time)
	ReversePlayback bool `json:"reversePlayback"`
}

type ExportRequest struct {
	Clips       []ExportClip `json:"clips"`
	Audio       *AudioSpec   `json:"audio,omitempty"`
	AspectRatio string       `json:"aspectRatio,omitempty"`
	CropMode    string       `json:"cropMode,omitempty"`
}

type ExportResponse struct {
	ExportID string `json:"exportId"`
	URL      string `json:"url"`
	Status   string `json:"status"` // processing | done | error
	Error    string `json:"error"`
}

type ExportItem struct {
	Filename string    `json:"filename"`
	URL      string    `json:"url"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modTime"`
}

type AudioSpec struct {
	AssetID string  `json:"assetId"`
	Volume  float64 `json:"volume"`
}

type Server struct {
	addr        string
	dataDir     string
	uploadDir   string
	exportDir   string
	workDir     string
	assetsIndex map[string]Asset
	jobs        map[string]*ExportResponse
	assetsFile  string
}

func main() {
	s := &Server{
		addr:    ":" + envOr("PORT", envOr("ADDR", defaultAddr)[1:]),
		dataDir: envOr("DATA_DIR", filepath.Join(".", "data")),
	}
	s.uploadDir = filepath.Join(s.dataDir, "uploads")
	s.exportDir = filepath.Join(s.dataDir, "exports")
	s.workDir = filepath.Join(s.dataDir, "work")
	mustMkdirAll(s.uploadDir)
	mustMkdirAll(s.exportDir)
	mustMkdirAll(s.workDir)
	s.assetsIndex = map[string]Asset{}
	s.jobs = map[string]*ExportResponse{}
	s.assetsFile = filepath.Join(s.dataDir, "assets.json")
	s.loadAssets()

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/upload", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(256 << 20); err != nil { // 256MB
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		asset, err := s.saveUploadedFile(file, header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, asset)
	}))

	mux.HandleFunc("/api/assets", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		items := make([]Asset, 0, len(s.assetsIndex))
		for _, a := range s.assetsIndex {
			items = append(items, a)
		}
		writeJSON(w, items)
	}))

	// Delete uploaded asset
	mux.HandleFunc("/api/assets/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
		if id == "" {
			http.Error(w, "missing asset id", http.StatusBadRequest)
			return
		}
		a, ok := s.assetsIndex[id]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_ = os.Remove(filepath.Join(s.uploadDir, a.Filename))
		delete(s.assetsIndex, id)
		_ = s.persistAssets()
		w.WriteHeader(http.StatusNoContent)
	}))

	mux.HandleFunc("/api/export", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := uuid.New().String()
		outPath := filepath.Join(s.exportDir, fmt.Sprintf("export-%s.mp4", id))
		job := &ExportResponse{ExportID: id, Status: "processing"}
		s.jobs[id] = job
		go func() {
			if err := s.concatenateClips(req, outPath); err != nil {
				log.Printf("export failed: %v", err)
				job.Status = "error"
				job.Error = err.Error()
				return
			}
			job.Status = "done"
			job.URL = "/exports/" + filepath.Base(outPath)
		}()
		writeJSON(w, job)
	}))

	// Export status polling
	mux.HandleFunc("/api/export/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/export/")
		if id == "" {
			http.Error(w, "missing export id", http.StatusBadRequest)
			return
		}
		job, ok := s.jobs[id]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, job)
	}))

	// List exports
	mux.HandleFunc("/api/exports", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		entries, err := os.ReadDir(s.exportDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items := make([]ExportItem, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".mp4") {
				continue
			}
			info, _ := e.Info()
			items = append(items, ExportItem{Filename: name, URL: "/exports/" + name, Size: info.Size(), ModTime: info.ModTime()})
		}
		// newest first
		for i := 0; i < len(items); i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].ModTime.Before(items[j].ModTime) {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
		writeJSON(w, items)
	}))

	// Delete an export by filename
	mux.HandleFunc("/api/exports/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/api/exports/")
		if name == "" {
			http.Error(w, "missing filename", http.StatusBadRequest)
			return
		}
		path := filepath.Join(s.exportDir, name)
		if err := os.Remove(path); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	// Static: serve uploaded files and exports
	mux.Handle("/uploads/", withCORSHandler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir)))))
	mux.Handle("/exports/", withCORSHandler(http.StripPrefix("/exports/", http.FileServer(http.Dir(s.exportDir)))))

	log.Printf("server listening on %s", s.addr)
	if err := http.ListenAndServe(s.addr, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) saveUploadedFile(file multipart.File, header *multipart.FileHeader) (Asset, error) {
	id := uuid.New().String()
	sanitized := sanitizeFilename(header.Filename)
	dst := filepath.Join(s.uploadDir, fmt.Sprintf("%s-%s", id, sanitized))
	out, err := os.Create(dst)
	if err != nil {
		return Asset{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return Asset{}, err
	}
	kind := inferKindByExt(dst)
	asset := Asset{
		ID:          id,
		Filename:    filepath.Base(dst),
		URL:         "/uploads/" + filepath.Base(dst),
		Kind:        kind,
		DurationSec: 0,
	}
	s.assetsIndex[id] = asset
	_ = s.persistAssets()
	return asset, nil
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "..", ".")
	return name
}

func inferKindByExt(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".mov", ".m4v", ".webm", ".mkv":
		return "video"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	case ".mp3", ".wav", ".m4a", ".aac", ".flac", ".ogg":
		return "audio"
	default:
		return "unknown"
	}
}

// concatenateClips builds a unified mp4 using ffmpeg via intermediate normalized segments.
func (s *Server) concatenateClips(req ExportRequest, outPath string) error {
	// Defer to ffmpeg service implementation in ffmpegservice.go
	return buildConcatenatedMP4(s, req, outPath)
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

// withCORSHandler wraps a generic http.Handler with CORS headers
func withCORSHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustMkdirAll(path string) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		log.Fatalf("failed to create dir %s: %v", path, err)
	}
}

// Helpers for creating deterministic segment file names.
func segmentFilename(ix int) string {
	return fmt.Sprintf("segment-%03d.mp4", ix)
}

// Unused but kept for potential future: returns a timestamped name.
func tsName(prefix, ext string) string {
	return fmt.Sprintf("%s-%d%s", prefix, time.Now().UnixNano(), ext)
}

// Persistence helpers for assets index
func (s *Server) loadAssets() {
	b, err := os.ReadFile(s.assetsFile)
	if err != nil {
		return
	}
	var arr []Asset
	if err := json.Unmarshal(b, &arr); err != nil {
		log.Printf("failed to read assets.json: %v", err)
		return
	}
	for _, a := range arr {
		s.assetsIndex[a.ID] = a
	}
}

func (s *Server) persistAssets() error {
	items := make([]Asset, 0, len(s.assetsIndex))
	for _, a := range s.assetsIndex {
		items = append(items, a)
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.assetsFile, b, 0o644)
}
