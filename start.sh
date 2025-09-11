#!/bin/bash

# Video Editor - Unified Start Script
echo "🎬 Starting Video Editor..."

# Kill any existing processes on our ports
echo "🧹 Cleaning up existing processes..."
lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null
lsof -ti tcp:5173 | xargs -r kill -9 2>/dev/null
sleep 1

# Function to cleanup processes on exit
cleanup() {
    echo "🛑 Shutting down servers..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    exit 0
}

# Trap SIGINT (Ctrl+C) and SIGTERM
trap cleanup SIGINT SIGTERM

# Start backend
echo "🚀 Starting backend server..."
cd server
go mod tidy
ADDR=:8080 go run . &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Start frontend
echo "🎨 Starting frontend server..."
cd ../web
npm install --silent
VITE_BACKEND_BASE=http://localhost:8080 npm run dev &
FRONTEND_PID=$!

echo "✅ Both servers started!"
echo "📱 Frontend: http://localhost:5173"
echo "🔧 Backend: http://localhost:8080"
echo "Press Ctrl+C to stop both servers"

# Wait for either process to exit
wait
