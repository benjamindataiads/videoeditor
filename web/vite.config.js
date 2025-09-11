import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    strictPort: true
  },
  define: {
    'import.meta.env.VITE_BACKEND_BASE': JSON.stringify(process.env.VITE_BACKEND_BASE || 'http://localhost:8080')
  }
})


