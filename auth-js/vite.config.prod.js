import { defineConfig } from 'vite'
import fs from 'fs'
import path from 'path'

export default defineConfig({
  root: './js',
  build: {
    outDir: '../dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    minify: false,//'esbuild',
    rollupOptions: {
      input: {
        main: './js/main.js'
      },
      output: {
        entryFileNames: 'assets/main-[hash].js',
        chunkFileNames: 'assets/chunks/[name]-[hash].js',
        assetFileNames: 'assets/main-[hash][extname]'
      }
    }
  },
  plugins: [
    {
      name: 'generate-manifest',
      closeBundle() {
        const distDir = path.resolve(__dirname, 'dist')
        const assetsDir = path.join(distDir, 'assets')
        const files = fs.readdirSync(assetsDir)
        
        const jsFile = files.find(f => f.startsWith('main-') && f.endsWith('.js'))
        const cssFile = files.find(f => f.startsWith('main-') && f.endsWith('.css'))

        if (!jsFile || !cssFile) {
          throw new Error('Could not find generated JS or CSS files')
        }

        const manifest = {
          js: `assets/${jsFile}`,
          css: `assets/${cssFile}`
        }

        // Write manifest to dist directory
        fs.writeFileSync(
          path.join(distDir, 'manifest.json'),
          JSON.stringify(manifest, null, 2)
        )

        // Clear the privous templates/static directory
        const templatesDir = path.resolve(__dirname, '../internal/web/auth/templates/static')
        if (fs.existsSync(templatesDir)) {
          fs.rmSync(templatesDir, { recursive: true })
        }

        // Copy assets to templates directory
        if (!fs.existsSync(templatesDir)) {
          fs.mkdirSync(templatesDir, { recursive: true })
        }

        // Copy manifest
        fs.copyFileSync(
          path.join(distDir, 'manifest.json'),
          path.join(templatesDir, 'manifest.json')
        )

        // Copy assets directory
        const templatesAssetsDir = path.join(templatesDir, 'assets')
        if (!fs.existsSync(templatesAssetsDir)) {
          fs.mkdirSync(templatesAssetsDir, { recursive: true })
        }

        // Copy all files from dist/assets to templates/static/assets
        fs.readdirSync(assetsDir).forEach(file => {
          fs.copyFileSync(
            path.join(assetsDir, file),
            path.join(templatesAssetsDir, file)
          )
        })

        console.log('Generated manifest:', manifest)
        console.log('Copied assets to templates directory')
      }
    }
  ]
})