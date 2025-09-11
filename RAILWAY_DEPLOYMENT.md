# Railway Deployment Guide for Video Editor

This guide will help you deploy your video editor app to Railway.

## Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **GitHub Repository**: Your code should be pushed to GitHub (already done: `https://github.com/benjamindataiads/videoeditor`)

## Deployment Steps

### Step 1: Connect to Railway

1. Go to [railway.app](https://railway.app) and sign in
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Choose your repository: `benjamindataiads/videoeditor`

### Step 2: Configure Backend Service

1. Railway will automatically detect your `railway.json` and deploy the backend
2. The backend will be available at a Railway-generated URL (e.g., `https://your-app-name.railway.app`)
3. Note this URL - you'll need it for the frontend configuration

### Step 3: Deploy Frontend (Separate Service)

1. In your Railway project dashboard, click "New Service"
2. Select "GitHub Repo" and choose the same repository
3. Configure the service:
   - **Root Directory**: `web`
   - **Build Command**: `npm run build`
   - **Start Command**: `serve -s dist -l $PORT`

### Step 4: Configure Environment Variables

#### Backend Service:
- No additional environment variables needed (Railway automatically provides `PORT`)

#### Frontend Service:
1. Go to your frontend service settings
2. Add environment variable:
   - **Name**: `VITE_BACKEND_BASE`
   - **Value**: `https://your-backend-service-name.railway.app` (replace with actual backend URL)

### Step 5: Update Frontend Environment File

Update `web/.env.production` with your actual Railway backend URL:

```bash
VITE_BACKEND_BASE=https://your-actual-backend-url.railway.app
```

### Step 6: Enable CORS (if needed)

If you encounter CORS issues, you may need to update your backend to allow requests from your frontend domain. The current backend should already handle CORS properly.

## Important Notes

### File Storage
- **Temporary files**: Railway's ephemeral filesystem means uploaded files and exports are temporary
- **For production**: Consider integrating with cloud storage (AWS S3, Cloudinary, etc.) for persistent file storage

### FFmpeg
- The backend Dockerfile includes FFmpeg installation
- Railway supports FFmpeg in Docker containers

### Domain Configuration
- Railway provides free subdomains for each service
- You can configure custom domains in Railway's dashboard if needed

## Troubleshooting

### Build Issues

#### Docker Build Timeout
If you see "context canceled" errors during Docker build:

1. **Try the optimized Dockerfile**: The main `Dockerfile` uses multi-stage builds for better performance
2. **Use Nixpacks instead**: Replace `railway.json` with `railway.nixpacks.json` and rename it to `railway.json`
3. **Use lightweight build**: Replace the main Dockerfile with `Dockerfile.lightweight` (removes FFmpeg from build)

#### Alternative Build Methods
```bash
# Method 1: Use Nixpacks (recommended for FFmpeg issues)
mv railway.json railway.dockerfile.json
mv railway.nixpacks.json railway.json

# Method 2: Use lightweight Docker (no FFmpeg in container)
mv server/Dockerfile server/Dockerfile.full
mv server/Dockerfile.lightweight server/Dockerfile
```

#### FFmpeg Installation Issues
- Railway's Docker builder sometimes times out on FFmpeg installation
- Nixpacks handles system dependencies better than Docker for this use case
- If FFmpeg is not available, video processing will fail but the app will still start

### Runtime Issues
- Check service logs in Railway dashboard
- Verify environment variables are set correctly
- Ensure frontend is pointing to correct backend URL
- Test endpoints manually: `curl https://your-app.railway.app/api/assets`

### CORS Issues
- Verify `VITE_BACKEND_BASE` environment variable is correct
- Check that backend allows requests from frontend domain
- Ensure both services are deployed and running

### Performance Issues
- Railway's free tier has resource limits
- Large video files may cause timeouts
- Consider upgrading Railway plan for production use

## Post-Deployment

1. Test file upload functionality
2. Test video export functionality
3. Verify all features work as expected
4. Monitor usage and performance in Railway dashboard

## Scaling Considerations

- Railway auto-scales based on usage
- For high-volume usage, consider:
  - Implementing cloud storage for files
  - Adding Redis for session management
  - Optimizing video processing workflows

## Costs

- Railway offers a generous free tier
- Costs scale with usage (compute time, bandwidth, storage)
- Monitor usage in Railway dashboard
