import fs from 'fs-extra'
import path from 'path'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import { isRawImage, imageSize } from './utils'
import config from '../config'

export default async function processImage({ driver, addFinishedImage }, id) {
  const session = driver.session()

  const result = await session.run(`MATCH (p:Photo { id: {id} }) RETURN p`, {
    id,
  })

  await session.run(
    `MATCH (p:Photo { id: {id} })-[rel]->(url:PhotoURL) DELETE url, rel`,
    { id }
  )

  const photo = result.records[0].get('p').properties

  console.log('Processing photo', photo.path)

  const imagePath = path.resolve(config.cachePath, 'images', id)

  await fs.remove(imagePath)
  await fs.mkdirp(imagePath)

  let originalPath = photo.path

  if (await isRawImage(photo.path)) {
    // console.log('Processing RAW image')

    const extractedPath = path.resolve(imagePath, 'extracted.jpg')
    await exiftool.extractPreview(photo.path, extractedPath)

    originalPath = extractedPath
  }

  // Resize image
  const thumbnailPath = path.resolve(imagePath, 'thumbnail.jpg')
  await sharp(originalPath)
    .jpeg({ quality: 80 })
    .resize(1440, 1080, { fit: 'inside', withoutEnlargement: true })
    .toFile(thumbnailPath)

  try {
    const { width: originalWidth, height: originalHeight } = await imageSize(
      originalPath
    )
    const { width: thumbnailWidth, height: thumbnailHeight } = await imageSize(
      thumbnailPath
    )

    await session.run(
      `MATCH (p:Photo { id: {id} })
    CREATE (thumbnail:PhotoURL { url: {thumbnailUrl}, width: {thumbnailWidth}, height: {thumbnailHeight} })
    CREATE (original:PhotoURL { url: {originalUrl}, width: {originalWidth}, height: {originalHeight} })
    CREATE (p)-[:THUMBNAIL_URL]->(thumbnail)
    CREATE (p)-[:ORIGINAL_URL]->(original)
    `,
      {
        id,
        thumbnailUrl: `/images/${id}/${path.basename(thumbnailPath)}`,
        thumbnailWidth,
        thumbnailHeight,
        originalUrl: `/images/${id}/${path.basename(originalPath)}`,
        originalWidth,
        originalHeight,
      }
    )
  } catch (e) {
    console.log('Create photo url failed', e)
  }

  session.close()

  addFinishedImage()
}
