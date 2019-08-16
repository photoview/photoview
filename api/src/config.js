import path from 'path'

export default {
  cachePath: process.env.PHOTO_CACHE || path.resolve(__dirname, 'cache'),
  host: new URL(process.env.API_ENDPOINT || 'http://localhost:4001/'),
}
