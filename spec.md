## Video Editor – Product and Engineering Specification

### 1) Purpose and scope
Web-based lightweight video editor for stitching images and videos, adding an optional background audio track, basic visual effects per clip (horizontal mirror, reverse playback), and transitions between adjacent clips. Output: a single MP4 (H.264/AAC) with selectable aspect ratio and crop/letterbox behavior. The app is single-user, sessionless, file-backed on disk (uploads/exports), deployable to a single container.

This document is intended for a senior engineer to rebuild the product from scratch with functional parity to the current app.

### 2) Personas and goals
- Creator (single user):
  - Upload media (videos, images, audio)
  - Compose a timeline with clips and an optional audio bed
  - Trim clips, set durations for images
  - Apply per-clip effects: horizontal mirror (hflip), reverse playback
  - Apply transitions (e.g., fade/dissolve/wipe/slide) between clips
  - Choose aspect ratio and crop/letterbox fit
  - Preview and export final MP4

### 3) High-level UX
- Left sidebar tabs: Media, Audio, FX (Transitions)
  - Media: upload images/videos, click to add to timeline
  - Audio: upload audio, click to add as single-track background
- Center: Player with aspect-ratio frame; play/stop
- Bottom: Timeline with clips in order, draggable trim handles, per-clip action buttons (mirror, reverse), and transition pickers between clips (list of transitions (Fade, Dissolve, Wipes, Slides, Circle, Radial)
- Right sidebar: Exports list with links to download or remove

### 4) Functional requirements (user stories)
1. Upload medi
   - As a user, I can upload images/videos (and audio) and see them in the sidebar with thumbnails/filenames
2. Build timeline
   - As a user, I can add media to timeline in order (append)
   - As a user, I can trim video clips (start/end), and set image clip duration
   - As a user, I can remove clips
3. Per-clip effects
   - Mirror: horizontal flip visual in preview and applied in export
   - Reverse playback: preview plays backward; export uses ffmpeg reverse filter
4. Transitions between clips
   - As a user, I can select a transition in FX and add it between two adjacent clips via a “+” hotspot
   - As a user, I can remove a transition badge (FX) between clips
   - Supported transitions (via ffmpeg xfade): fade, dissolve, wipeleft, wiperight, slideleft, slideright, circlecrop, radial
5. Global settings
   - Aspect ratio: 16:9 (default), 1:1, 9:16
   - Fit: crop to fill (center crop) or letterbox/pad to fit (black)
6. Background audio track
   - Optional single audio asset with volume (default 1.0)
   - Trimmed to final video duration in export
7. Preview
   - Play/stop timeline; player always shows playhead advancing left→right regardless of reverse effect
   - Reverse preview simulated by stepping currentTime backward at ~30 fps
   - Transition preview simulated by overlaying the next clip (video or image) with opacity ramp during the last transition seconds of current clip
8. Export
   - As a user, I can export; the server merges clips (+ audio) and returns an MP4 URL in an async job. Exports are listed with size/time, can be downloaded or deleted

### 5) Non-functional requirements
- Target: desktop browsers (Chromium/WebKit/Gecko latest)
- No authentication required
- Files stored on local disk (uploads, exports)
- Export must complete for typical short videos (< 2 minutes total) on a single CPU core within reasonable time (minutes)

### 6) Architecture overview
- Frontend: Vue 3 + Vite + Tailwind CSS (single page, `web/`)
- Backend: Go 1.21+ HTTP server (no frameworks) + ffmpeg via `github.com/u2takey/ffmpeg-go` wrapper (in `server/`)
- Storage: disk directories
  - `server/data/uploads` – user uploads (originals)
  - `server/data/work` – normalized per-export segments
  - `server/data/exports` – final MP4 files
  - `server/data/assets.json` – index of uploaded assets

### 7) Data model (backend)
- Asset
  - id: string (UUID)
  - filename: string (server filename)
  - url: string (/uploads/<filename>)
  - kind: "image" | "video" | "audio"
  - durationSec: float64 (optional, not relied upon for normalization)
- ExportClip
  - assetId: string
  - startSec: float64 (video trim start)
  - endSec: float64 (video trim end)
  - durationSec: float64 (image duration)
  - reversed: bool (horizontal mirror)
  - reversePlayback: bool (play clip backward)
- AudioSpec { assetId: string, volume: float64 }
- Transition { clipIndex: int, transitionId: string, duration: float64 }
- ExportRequest
  - clips: []ExportClip
  - audio?: AudioSpec
  - aspectRatio: "16:9" | "1:1" | "9:16"
  - cropMode: "letterbox" | "crop"
  - transitions?: []Transition
- ExportResponse { exportId, url, status: processing|done|error, error? }

### 8) HTTP API (backend)
- POST /api/upload multipart/form-data (file)
  - returns Asset
- GET /api/assets → []Asset
- DELETE /api/assets/:id → 204
- POST /api/export → ExportResponse (processing or done)
  - body: ExportRequest JSON
- GET /api/export/:id → ExportResponse
- GET /api/exports → []ExportItem { filename,url,size,modTime }
- DELETE /api/exports/:filename → 204
- Static: /uploads/*, /exports/* served from disk

### 9) Frontend application behavior
- App state
  - assets[], assetDurations{id→duration}, clips[], audioClips[], transitions{clipIndex→{id,duration}}
  - aspectRatio, cropMode
  - player refs: player (video), audioPlayer, overlayPlayer
  - timeline geometry: pxPerSec = 80, trackLeftPad = 20
- Timeline
  - displayDuration(clip):
    - if video: (endSec - startSec) with fallback to probed duration
    - if image: durationSec (default 1.0)
  - trim handles update start/end or duration; values clamped with 0.2s min
- Preview effects
  - Mirror: apply CSS `scale-x-[-1]` on player and clip thumbnails
  - Reverse: step currentTime backward via setInterval at 30 fps; playhead continues forward (computed from elapsed, not video time)
  - Transition preview: 
    - When a transition exists after clip i with duration d, during last d seconds of clip i:
      - If next clip is video: overlay <video> with src of clip i+1, start at its startSec, opacity ramp 0→1
      - If image: overlay <img> with opacity ramp
    - Overlay uses same object-fit class as the base player
- Export request payload includes clips, audio, aspectRatio, cropMode, transitions[]

### 10) Export pipeline (server) – ffmpeg
1) Normalize each clip into a segment MP4 (H.264 yuv420p, 30 fps) with consistent frame size and SAR
   - Aspect ratio target by request
     - crop mode:
       - `scale=WxH:force_original_aspect_ratio=increase,crop=W:H,setsar=1`
     - letterbox mode:
       - `scale=WxH:force_original_aspect_ratio=decrease,pad=W:H:(W-iw)/2:(H-ih)/2:black,setsar=1`
   - Mirror: append `,hflip` to vf chain
   - Reverse playback:
     - for video: build `... ,trim=start=..:end=..,setpts=PTS-STARTPTS,reverse[,hflip]`
     - for image: not applicable (handled as segment duration)
   - Video trims use input `-ss` and output `-t` when not reversed
   - Image: loop to duration `-t <seconds>`
   - Record `segDurations[i]` for each segment (end-start for video; duration for image; fallback 1.0)
2) Chain segments with transitions
   - Build iterative chain `out` starting from seg0
   - For boundary after seg i:
     - if transition exists t (type, duration d):
       - offset for xfade = cumulativeTimelineSoFar − d (i.e., duration(seg0..i) − d), clamped to ≥ 0 (per FFmpeg Xfade guidance)
       - `out = xfade(out, seg{i+1}, transition=<type>, duration=d, offset=<computed>)`
       - update cumulative = cumulative + segDurations[i+1] − d
     - else: `out = concat(out, seg{i+1})`; cumulative += segDurations[i+1]
   - Supported transitions map to xfade names: fade, dissolve, wipeleft, wiperight, slideleft, slideright, circlecrop, radial
3) Optional audio
   - If audio provided, load and adjust volume
   - Compute total video duration as sum(segDurations) minus total overlapped transition time (sum of d for each transition applied)
   - `atrim=0:total` and mux with video
4) Output final MP4 with `libx264`, `yuv420p`, `r=30`

### 11) Edge cases and rules
- Clip minimum visual duration 0.2s (UI clamp)
- Reverse + mirror can be combined
- Transition duration must be ≤ duration of the preceding segment; if larger, clamp or set offset=0 to begin immediately
- If next media is image, xfade still works (image segment is a real video segment with duration); preview uses image overlay
- If any segment duration is unknown, assume 1.0s to avoid pipeline failure

### 12) Error handling
- Backend returns 4xx for invalid input (missing assets, malformed JSON)
- Export job: status error with message on ffmpeg failure
- Frontend shows generic error toast for export failure

### 13) Build & deploy
- Frontend: `npm run build` (Vite) → static files served by separate hosting or the Go server in production if desired
- Backend: single Go binary; ffmpeg must be available in image/base layer; container mounts persistent volume for `data/`

### 14) Acceptance criteria (critical)
- Upload image/video/audio; media appear in sidebars; timeline allows adding clips
- Trim video, set image duration; play preview works
- Mirror effect visible in preview and in export
- Reverse playback visible in preview (content backwards) and in export
- FX transitions visible in preview overlay (video→video, video→image, image→video, image→image) and in export
- Export produces MP4 with correct aspect ratio, crop/letterbox behavior, and optional audio bed trimmed to final length

### 15) Future improvements (not in MVP)
- Drag-and-drop clip reordering; multiple audio tracks; per-clip volume; per-clip transitions pickers; GPU filters; thumbnails scrub; keyboard shortcuts; multi-user auth and cloud storage; worker queue for exports


