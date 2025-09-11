# Video Editor MVP

A modern web-based video editor built with Go + Vue.js + Tailwind CSS. Upload images/videos/audio, arrange clips on a timeline with trim controls, and export merged MP4s with customizable aspect ratios.

## Features

- **Multi-format support**: Images (PNG, JPG), videos (MP4, MOV), audio (MP3, WAV)
- **Interactive timeline**: Drag to trim clips, visual playhead, hover preview
- **Aspect ratio control**: 16:9, 1:1, 9:16 with fit/crop options
- **Audio mixing**: Add background audio tracks synchronized with video
- **Real-time preview**: See exactly how your export will look
- **Export management**: Download and manage exported videos

## Requirements

- **Go 1.22+**
- **Node.js 18+** 
- **FFmpeg** installed and accessible in PATH (`ffmpeg` and `ffprobe`)

## Quick Start

### Option 1: Unified Start (Recommended)

```bash
./start.sh
```

This launches both backend and frontend automatically. Press `Ctrl+C` to stop both servers.

### Option 2: Manual Start

**Terminal 1 (Backend):**
```bash
cd server
go mod tidy
ADDR=:8080 go run .
```

**Terminal 2 (Frontend):**
```bash
cd web
npm install
VITE_BACKEND_BASE=http://localhost:8080 npm run dev
```

### Access the App

Visit **http://localhost:5173** in your browser.

## Usage

1. **Upload assets**: Use "Upload Media" for videos/images, "Upload Audio" for audio tracks
2. **Build timeline**: Click assets to add them to the timeline; drag handles to trim
3. **Configure output**: Select aspect ratio (16:9/1:1/9:16) and fit mode (Fit/Crop)
4. **Preview**: Use the play button to preview your timeline
5. **Export**: Click Export to generate MP4; download from the right sidebar

## API Endpoints

- `POST /api/upload` - Upload media files
- `GET /api/assets` - List uploaded assets
- `DELETE /api/assets/:id` - Delete an asset
- `POST /api/export` - Start export job
- `GET /api/export/:id` - Check export status
- `GET /api/exports` - List completed exports
- `DELETE /api/exports/:filename` - Delete an export

## Technical Details

- **Backend**: Go with ffmpeg-go for video processing
- **Frontend**: Vue 3 + Vite + Tailwind CSS + Heroicons
- **Video processing**: All clips normalized and concatenated via FFmpeg
- **Audio sync**: Background audio trimmed to match video timeline duration
- **Aspect ratios**: Dynamic scaling with letterboxing or cropping as needed


