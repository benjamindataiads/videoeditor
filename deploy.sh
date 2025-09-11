#!/bin/bash

# Video Editor - Railway Deployment Helper
echo "🚀 Preparing for Railway deployment..."

# Check if we're in the right directory
if [ ! -f "railway.json" ]; then
    echo "❌ Error: railway.json not found. Make sure you're in the project root directory."
    exit 1
fi

echo "📋 Pre-deployment checklist:"
echo "✅ Dockerfiles created"
echo "✅ Railway configuration ready"
echo "✅ Environment files configured"

echo ""
echo "🔧 Next steps:"
echo "1. Push your changes to GitHub:"
echo "   git add ."
echo "   git commit -m 'Add Railway deployment configuration'"
echo "   git push origin main"
echo ""
echo "2. Go to https://railway.app and create a new project"
echo "3. Connect your GitHub repository: benjamindataiads/videoeditor"
echo "4. Railway will automatically deploy your backend using railway.json"
echo "5. Create a second service for the frontend (see RAILWAY_DEPLOYMENT.md)"
echo ""
echo "📚 For detailed instructions, see RAILWAY_DEPLOYMENT.md"
echo ""
echo "🎬 Your video editor will be live soon!"
