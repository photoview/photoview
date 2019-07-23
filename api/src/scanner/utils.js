import fs from 'fs-extra'
import readChunk from 'read-chunk'
import imageType from 'image-type'
import { promisify } from 'util'
import path from 'path'
import config from '../config'

export const isImage = async path => {
  if ((await fs.stat(path)).isDirectory()) {
    return false
  }

  try {
    const buffer = await readChunk(path, 0, 12)
    const type = imageType(buffer)
    return type != null
  } catch (e) {
    throw new Error(`isImage error at ${path}: ${JSON.stringify(e)}`)
  }
}

export const isRawImage = async path => {
  try {
    const buffer = await readChunk(path, 0, 12)
    const { ext } = imageType(buffer)

    const rawTypes = ['cr2', 'arw', 'crw', 'dng']

    return rawTypes.includes(ext)
  } catch (e) {
    throw new Error(`isRawImage error at ${path}: ${JSON.stringify(e)}`)
  }
}

export const imageSize = promisify(require('image-size'))

export const getImageCachePath = id =>
  path.resolve(config.cachePath, 'images', id)
