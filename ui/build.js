const fs = require('fs-extra')
const esbuild = require('esbuild')
const bs = require('browser-sync').create()
const historyApiFallback = require('connect-history-api-fallback')

require('dotenv').config()

const production = process.env.NODE_ENV == 'production'
const watchMode = process.argv[2] == 'watch'

const ENVIRONMENT_VARIABLES = ['NODE_ENV', 'PHOTOVIEW_API_ENDPOINT']

const defineEnv = ENVIRONMENT_VARIABLES.reduce((acc, key) => {
  acc[`process.env.${key}`] = `"${process.env[key]}"`
  return acc
}, {})

const esbuildOptions = {
  entryPoints: ['src/index.js'],
  publicPath: process.env.UI_PUBLIC_URL || '/',
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
  define: defineEnv,
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

  require('workbox-build').generateSW({
    globDirectory: 'dist/',
    globPatterns: ['**/*.{png,svg,woff2,ttf,eot,woff,js,ico,html,json,css}'],
    swDest: 'dist/service-worker.js',
  })
}
