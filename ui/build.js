const fs = require('fs')
const bs = require('browser-sync').create()
const historyApiFallback = require('connect-history-api-fallback')

const production = process.env.NODE_ENV == 'production'

let builderPromise = require('esbuild').build({
  entryPoints: ['src/index.js'],
  outdir: 'dist',
  // format: 'esm',
  bundle: true,
  platform: 'browser',
  target: ['chrome58', 'firefox57', 'safari11', 'edge16'],
  // splitting: true,
  minify: production,
  sourcemap: !production,
  loader: {
    '.js': 'jsx',
    '.svg': 'text',
    '.woff': 'file',
    '.woff2': 'file',
    '.ttf': 'file',
    '.eot': 'file',
    '.png': 'file',
  },
  define: {
    'process.env.NODE_ENV': '"development"',
  },
  incremental: true,
})

fs.copyFileSync('src/index.html', 'dist/index.html')

bs.init({
  server: {
    baseDir: './dist',
    middleware: [historyApiFallback()],
  },
  port: 1234,
  open: false,
})

bs.watch('src/**/*.js').on('change', async args => {
  console.log('reloading', args)
  builderPromise = (await builderPromise).rebuild()
  bs.reload(args)
})
