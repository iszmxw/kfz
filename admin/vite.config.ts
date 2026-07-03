import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  base: '/admin/',
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/admin/api/v1': {
        target: 'http://127.0.0.1:8888',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist'
  }
});
