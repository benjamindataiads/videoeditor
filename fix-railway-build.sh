#!/bin/bash

echo "🔧 Fixing Railway build issues..."

# Check if we're in the right directory
if [ ! -f "railway.json" ]; then
    echo "❌ Error: railway.json not found. Make sure you're in the project root directory."
    exit 1
fi

echo "📋 Available fixes:"
echo "1. Switch to Nixpacks (recommended for FFmpeg issues)"
echo "2. Use lightweight Docker build (no FFmpeg)"
echo "3. Keep current Docker build (optimized)"

read -p "Choose option (1-3): " choice

case $choice in
    1)
        echo "🔄 Switching to Nixpacks build..."
        mv railway.json railway.dockerfile.json
        mv railway.nixpacks.json railway.json
        echo "✅ Switched to Nixpacks. Nixpacks will handle FFmpeg installation."
        echo "📝 Make sure server/nixpacks.toml is committed to git."
        ;;
    2)
        echo "🔄 Switching to lightweight Docker build..."
        mv server/Dockerfile server/Dockerfile.full
        mv server/Dockerfile.lightweight server/Dockerfile
        echo "✅ Switched to lightweight build. Video processing may not work without FFmpeg."
        echo "⚠️  Consider using option 1 (Nixpacks) for full functionality."
        ;;
    3)
        echo "📋 Current optimized Docker build is already configured."
        echo "💡 If builds still fail, try option 1 or 2."
        ;;
    *)
        echo "❌ Invalid option. Exiting."
        exit 1
        ;;
esac

echo ""
echo "🚀 Next steps:"
echo "1. Commit and push changes:"
echo "   git add ."
echo "   git commit -m 'Fix Railway build configuration'"
echo "   git push origin main"
echo "2. Redeploy on Railway (it should auto-deploy from git push)"
echo "3. Check build logs in Railway dashboard"

echo ""
echo "✅ Railway build fix applied!"
