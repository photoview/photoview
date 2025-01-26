/// <reference types="vitest" />
/// <reference types="vite/client" />

import { defineConfig } from 'vite'
import svgr from 'vite-plugin-svgr'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    svgr(),
    VitePWA({
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'service-worker.ts',
      injectRegister: 'auto',
      manifest: false,
      injectManifest: {
        injectionPoint: undefined
      }
    })],
  envPrefix: ['VITE_', 'REACT_APP_'],
  server: {
    port: 1234,
  },
  esbuild: {
    logOverride: { 'this-is-undefined-in-esm': 'silent' },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './testing/setupTests.ts',
    reporters: ['verbose', 'junit'],
    outputFile: {
      junit: './junit-report.xml',
    },
    coverage: {
      reporter: ['text', 'lcov', 'json', 'html'],
    },
  },
})
