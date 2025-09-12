package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// buildConcatenatedMP4 normalizes all clips to h264/aac mp4 segments and concatenates them with transitions.
func buildConcatenatedMP4(s *Server, req ExportRequest, outPath string) error {
	// Prepare work directory for this job
	jobDir := filepath.Join(s.workDir, tsName("job", ""))
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		return err
	}

	// 1) Build normalized segments
	var segPaths []string
	for i, clip := range req.Clips {
		segPath := filepath.Join(jobDir, segmentFilename(i))
		asset, ok := s.assetsIndex[clip.AssetID]
		if !ok {
			return fmt.Errorf("unknown asset: %s", clip.AssetID)
		}
		inputPath := filepath.Join(s.uploadDir, asset.Filename)

		// Determine target dimensions based on aspect ratio
		var width, height int
		switch req.AspectRatio {
		case "1:1":
			width, height = 720, 720
		case "9:16":
			width, height = 720, 1280
		default: // "16:9"
			width, height = 1280, 720
		}

		var vfChain string
		if req.CropMode == "crop" {
			// Crop to fill - may cut content but fills frame completely
			vfChain = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,setsar=1",
				width, height, width, height)
		} else {
			// Letterbox - fit content with padding (default)
			vfChain = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(%d-iw)/2:(%d-ih)/2:black,setsar=1",
				width, height, width, height, width, height)
		}

		// Add horizontal flip if reversed (for normal playback)
		if clip.Reversed {
			vfChain += ",hflip"
		}

		if asset.Kind == "image" {
			// Turn still image into short H.264 segment
			if clip.DurationSec <= 0 {
				clip.DurationSec = 1
			}
			err := ffmpeg.Input(inputPath, ffmpeg.KwArgs{"loop": 1}).
				Output(segPath,
					ffmpeg.KwArgs{
						"r":       30,
						"t":       fmt.Sprintf("%f", clip.DurationSec),
						"pix_fmt": "yuv420p",
						"vcodec":  "libx264",
						"vf":      vfChain,
					},
				).
				OverWriteOutput().
				Run()
			if err != nil {
				return err
			}
		} else { // video
			start := clip.StartSec
			end := clip.EndSec
			dur := 0.0
			if end > 0 && end > start {
				dur = end - start
			}

			if clip.ReversePlayback {
				// For reverse playback, we need to handle trimming differently
				// First apply trimming in the filter chain, then reverse
				var trimFilter string
				if start > 0 || (end > 0 && end > start) {
					if end > 0 && end > start {
						trimFilter = fmt.Sprintf("trim=start=%f:end=%f,setpts=PTS-STARTPTS", start, end)
					} else if start > 0 {
						trimFilter = fmt.Sprintf("trim=start=%f,setpts=PTS-STARTPTS", start)
					}
				}

				// Build the complete filter chain: scale -> trim -> reverse -> hflip (if needed)
				var fullVfChain string
				if req.CropMode == "crop" {
					fullVfChain = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,setsar=1",
						width, height, width, height)
				} else {
					fullVfChain = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(%d-iw)/2:(%d-ih)/2:black,setsar=1",
						width, height, width, height, width, height)
				}

				if trimFilter != "" {
					fullVfChain += "," + trimFilter
				}
				fullVfChain += ",reverse"
				if clip.Reversed {
					fullVfChain += ",hflip"
				}

				in := ffmpeg.Input(inputPath)
				if err := in.Output(segPath, ffmpeg.KwArgs{
					"vcodec":  "libx264",
					"r":       30,
					"pix_fmt": "yuv420p",
					"vf":      fullVfChain,
				}).OverWriteOutput().Run(); err != nil {
					return err
				}
			} else {
				// Normal forward playback - use input seek for efficiency
				inKw := ffmpeg.KwArgs{}
				if start > 0 {
					inKw["ss"] = fmt.Sprintf("%f", start)
				}
				in := ffmpeg.Input(inputPath, inKw)
				outKw := ffmpeg.KwArgs{
					"vcodec":  "libx264",
					"r":       30,
					"pix_fmt": "yuv420p",
					"vf":      vfChain,
				}
				if dur > 0 {
					outKw["t"] = fmt.Sprintf("%f", dur)
				}
				if err := in.Output(segPath, outKw).OverWriteOutput().Run(); err != nil {
					return err
				}
			}
		}
		segPaths = append(segPaths, segPath)
	}

	// 2) Build video chain with transitions
	var videoOut *ffmpeg.Stream
	if len(segPaths) == 0 {
		return fmt.Errorf("no segments to export")
	}

	if len(segPaths) == 1 {
		// Single clip, no transitions needed
		videoOut = ffmpeg.Input(segPaths[0]).Video()
	} else {
		// Multiple clips, apply transitions
		videoOut = buildVideoChainWithTransitions(segPaths, req.Transitions)
	}

	// Optional audio
	var audioIn *ffmpeg.Stream
	if req.Audio != nil {
		a, ok := s.assetsIndex[req.Audio.AssetID]
		if !ok {
			return fmt.Errorf("unknown audio asset: %s", req.Audio.AssetID)
		}
		aPath := filepath.Join(s.uploadDir, a.Filename)
		audioIn = ffmpeg.Input(aPath)
		if req.Audio.Volume > 0 && req.Audio.Volume != 1 {
			audioIn = audioIn.Filter("volume", ffmpeg.Args{fmt.Sprintf("%f", req.Audio.Volume)})
		}
	}

	if audioIn != nil {
		// Calculate total video duration to trim audio if needed
		totalDur := 0.0
		for _, clip := range req.Clips {
			if clip.DurationSec > 0 {
				totalDur += clip.DurationSec
			} else {
				totalDur += 1.0 // default for images
			}
		}

		// Trim audio to match video timeline duration
		audioTrimmed := audioIn
		if totalDur > 0 {
			audioTrimmed = audioIn.Filter("atrim", ffmpeg.Args{fmt.Sprintf("0:%f", totalDur)})
		}

		// Mix video and audio
		if err := ffmpeg.Output([]*ffmpeg.Stream{videoOut, audioTrimmed}, outPath, ffmpeg.KwArgs{
			"c:v": "libx264", "pix_fmt": "yuv420p", "r": 30, "c:a": "aac",
		}).OverWriteOutput().Run(); err != nil {
			return err
		}
	} else {
		// Video only
		if err := videoOut.Output(outPath, ffmpeg.KwArgs{
			"c:v": "libx264", "pix_fmt": "yuv420p", "r": 30,
		}).OverWriteOutput().Run(); err != nil {
			return err
		}
	}
	return nil
}

// buildVideoChainWithTransitions creates a complex filter chain with xfade transitions
func buildVideoChainWithTransitions(segPaths []string, transitions []Transition) *ffmpeg.Stream {
	if len(segPaths) <= 1 {
		if len(segPaths) == 1 {
			return ffmpeg.Input(segPaths[0]).Video()
		}
		return nil
	}

	// Create a map of transitions by clip index for easy lookup
	transitionMap := make(map[int]Transition)
	for _, t := range transitions {
		transitionMap[t.ClipIndex] = t
	}

	// Start with the first clip
	result := ffmpeg.Input(segPaths[0]).Video()

	// Apply transitions between consecutive clips
	for i := 1; i < len(segPaths); i++ {
		nextClip := ffmpeg.Input(segPaths[i]).Video()

		// Check if there's a transition after the previous clip (index i-1)
		if transition, hasTransition := transitionMap[i-1]; hasTransition {
			// Apply xfade transition
			xfadeType := getXfadeTransition(transition.TransitionID)

			// For xfade to work properly, we need to calculate the offset
			// The offset should be set so the transition starts before the first video ends
			// This is a simplified calculation - in a real implementation you'd need to
			// track the actual durations and calculate precise offsets
			offset := 0.5 // Start transition 0.5 seconds before the end of current video

			result = ffmpeg.Filter([]*ffmpeg.Stream{result, nextClip}, "xfade", ffmpeg.Args{
				fmt.Sprintf("transition=%s", xfadeType),
				fmt.Sprintf("duration=%.2f", transition.Duration),
				fmt.Sprintf("offset=%.2f", offset),
			})
		} else {
			// No transition, just concatenate
			result = ffmpeg.Concat([]*ffmpeg.Stream{result, nextClip}, ffmpeg.KwArgs{"v": 1, "a": 0})
		}
	}

	return result
}

// getXfadeTransition maps our transition IDs to FFmpeg xfade transition names
func getXfadeTransition(transitionID string) string {
	switch transitionID {
	case "fade":
		return "fade"
	case "dissolve":
		return "dissolve"
	case "wipeleft":
		return "wipeleft"
	case "wiperight":
		return "wiperight"
	case "slideleft":
		return "slideleft"
	case "slideright":
		return "slideright"
	case "circlecrop":
		return "circlecrop"
	case "radial":
		return "radial"
	default:
		return "fade" // fallback to fade
	}
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep
		out += parts[i]
	}
	return out
}

// For debugging
func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := w.WriteString(l + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}
