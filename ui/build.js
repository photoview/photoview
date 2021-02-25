const fs = require('fs-extra')
const esbuild = require('esbuild')
const bs = require('browser-sync').create()
const historyApiFallback = require('connect-history-api-fallback')

const production = process.env.NODE_ENV == 'production'
const watchMode = process.argv[2] == 'watch'

const esbuildOptions = {
  entryPoints: ['src/index.js'],
  publicPath: '/',
  outdir: 'dist',
  format: 'esm',
  bundle: true,
  platform: 'browser',
  target: ['chrome58', 'firefox57', 'safari11', 'edge16'],
  splitting: true,
  minify: production,
  sourcemap: !production,
  loader: {
    '.js': 'jsx',
    '.svg': 'file',
    '.woff': 'file',
    '.woff2': 'file',
    '.ttf': 'file',
    '.eot': 'file',
    '.png': 'file',
  },
  define: {
    'process.env.PHOTOVIEW_API_ENDPOINT': '"http://localhost:4001/"',
    'process.env.NODE_ENV': '"development"',
  },
  incremental: watchMode,
}

fs.emptyDirSync('dist/')
fs.copyFileSync('src/index.html', 'dist/index.html')
fs.copyFileSync('src/manifest.webmanifest', 'dist/manifest.json')
fs.copyFileSync('src/favicon.ico', 'dist/favicon.ico')
fs.copySync('src/assets/', 'dist/assets/')

if (watchMode) {
  let builderPromise = esbuild.build(esbuildOptions)

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
} else {
  esbuild.buildSync(esbuildOptions)
}
