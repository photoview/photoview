import fs from 'fs-extra'
import path from 'path'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import { isRawImage, imageSize, getImageCachePath } from './utils'

export default async function processImage({ driver, addFinishedImage }, id) {
  const session = driver.session()

  const result = await session.run(
    `MATCH (p:Photo { id: {id} })<-[:CONTAINS]-(a:Album) RETURN p, a.id`,
    {
      id,
    }
  )

  await session.run(
    `MATCH (p:Photo { id: {id} })-->(urls:PhotoURL) DETACH DELETE urls`,
    { id }
  )

  const photo = result.records[0].get('p').properties
  const albumId = result.records[0].get('a.id')

  // console.log('Processing photo', photo.path)

  const imagePath = getImageCachePath(id, albumId)

  await fs.remove(imagePath)
  await fs.mkdirp(imagePath)

  let originalPath = photo.path

  if (await isRawImage(photo.path)) {
    // console.log('Processing RAW image')

    const extractedPath = path.resolve(imagePath, 'extracted.jpg')
    await exiftool.extractPreview(photo.path, extractedPath)

    const rawTags = await exiftool.read(photo.path)
    // ISO, FNumber, Model, ExposureTime, FocalLength, LensType
    // console.log(rawTags)

    let rotateAngle = null
    switch (rawTags.Orientation) {
      case 8:
        rotateAngle = -90
        break
      case 3:
        rotateAngle = 180
        break
      case 6:
        rotateAngle = 90
    }

    // Replace extension with .jpg
    let processedBase = path.basename(photo.path).match(/(.*)(\..*)/)
    processedBase =
      processedBase == null ? path.basename(photo.path) : processedBase[1]
    processedBase += '.jpg'

    const processedPath = path.resolve(imagePath, processedBase)
    await sharp(extractedPath)
      .jpeg({ quality: 80 })
      .rotate(rotateAngle)
      .toFile(processedPath)

    fs.remove(extractedPath)

    originalPath = processedPath
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
