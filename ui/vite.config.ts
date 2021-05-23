import { defineConfig } from 'vite'
import reactRefresh from '@vitejs/plugin-react-refresh'
import { babel } from '@rollup/plugin-babel'
import svgr from 'vite-plugin-svgr'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    reactRefresh(),
    // babel({
    //   babelHelpers: 'bundled',
    //   exclude: 'node_modules/**',
    //   extensions: ['.js', '.jsx', '.ts', '.tsx'],
    // }),
    svgr(),
  ],
  server: {
    port: 1234,
  },
})
