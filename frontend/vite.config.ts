import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    proxy: {
      // Proxy all /ws requests to the Go backend
      '/ws': {
        target: 'ws://localhost:8080', // Go backend WebSocket address
        ws: true, // Enable WebSocket proxy
      },
    },
  },
})
