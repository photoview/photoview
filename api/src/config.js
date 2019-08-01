import path from 'path'

export default {
  cachePath: path.resolve(__dirname, 'cache'),
  host: new URL(process.env.GRAPHQL_LISTEN_HOST || 'http://localhost:4001/'),
}
