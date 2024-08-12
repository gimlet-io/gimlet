import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import eslint from 'vite-plugin-eslint'

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    outDir: 'build',
    sourcemap: true,
  },
  base: '/',
  plugins: [react(), eslint()],
  server: {
    open: true,
    port: 3333,
    proxy: {
      '/api': 'http://localhost:9000',
      '/flags': 'http://localhost:9000',
    },
  },
})
