import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        // 127.0.0.1 (not localhost): avoids IPv6 ::1 resolution mismatch —
        // the Go server listens on IPv4 only.
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
})
