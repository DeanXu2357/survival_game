import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    proxy: {
      // 將所有指向 /ws 的請求，代理到您的 Go 後端
      '/ws': {
        target: 'ws://localhost:8080', // 您的 Go 後端 WebSocket 位址
        ws: true, // 啟用 WebSocket 代理
      },
    },
  },
})
