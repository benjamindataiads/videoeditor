import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    strictPort: true
  },
  define: {
    'import.meta.env.VITE_BACKEND_BASE': JSON.stringify(process.env.VITE_BACKEND_BASE || 'https://videoeditor-production-3bd0.up.railway.app')
  }
})


