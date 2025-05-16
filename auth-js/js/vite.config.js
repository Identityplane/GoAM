import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: '../dist',
    assetsDir: 'assets',
    rollupOptions: {
      input: {
        main: './main.js'
      },
      output: {
        entryFileNames: 'assets/main-[hash].js',
        chunkFileNames: 'assets/chunks/[name]-[hash].js',
        assetFileNames: 'assets/main-[hash][extname]'
      }
    }
  }
}); 