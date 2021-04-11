import fs from 'fs-extra'
import esbuild from 'esbuild'
import babel from 'esbuild-plugin-babel'
import browserSync from 'browser-sync'
import historyApiFallback from 'connect-history-api-fallback'
import dotenv from 'dotenv'
import workboxBuild from 'workbox-build'

dotenv.config()
const bs = browserSync.create()

const production = process.env.NODE_ENV == 'production'
const watchMode = process.argv[2] == 'watch'

const ENVIRONMENT_VARIABLES = ['NODE_ENV', 'PHOTOVIEW_API_ENDPOINT']

const defineEnv = ENVIRONMENT_VARIABLES.reduce((acc, key) => {
  acc[`process.env.${key}`] = process.env[key] ? `"${process.env[key]}"` : null
  return acc
}, {})

const esbuildOptions = {
  entryPoints: ['src/index.js'],
  plugins: [
    babel({
      filter: /photoview\/ui\/src\/.*\.js$/,
    }),
  ],
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
  const esbuildPromise = esbuild
    .build(esbuildOptions)
    .then(() => console.log('esbuild done'))

  const workboxPromise = workboxBuild
    .generateSW({
      globDirectory: 'dist/',
      globPatterns: ['**/*.{png,svg,woff2,ttf,eot,woff,js,ico,html,json,css}'],
      swDest: 'dist/service-worker.js',
    })
    .then(() => console.log('workbox done'))

  Promise.all([esbuildPromise, workboxPromise]).then(() =>
    console.log('build complete')
  )
}
