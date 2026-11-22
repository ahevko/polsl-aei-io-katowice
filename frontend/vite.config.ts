import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Serwer deweloperski Vite (port 5173) przekazuje żądania /api do backendu Go.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
