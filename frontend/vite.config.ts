import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Separate chunk for vega libraries to enable lazy loading
          'vega-libs': ['vega-lite', 'react-vega']
        }
      }
    }
  },
  optimizeDeps: {
    // Exclude vega libraries from pre-bundling to enable proper dynamic imports
    exclude: ['vega-lite', 'react-vega']
  }
})