import path from 'path'

// set environment variables from ../.env
require('dotenv').config()

let encryptionSaltRounds = 10
if (parseInt(process.env.ENCRYPTION_SALT_ROUNDS)) {
  encryptionSaltRounds = parseInt(process.env.ENCRYPTION_SALT_ROUNDS)
}

export default {
  cachePath: process.env.PHOTO_CACHE || path.resolve(__dirname, 'cache'),
  host: new URL(process.env.API_ENDPOINT || 'http://localhost:4001/'),
  encryptionSaltRounds,
}
