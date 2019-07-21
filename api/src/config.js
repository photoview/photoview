import path from 'path'

export default {
  cachePath: path.resolve(__dirname, 'cache'),
  host: process.env.HOST || 'http://localhost:4001/',
}
