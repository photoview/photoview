import fs from 'fs-extra'
import path from 'path'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import { isRawImage, imageSize, getImageCachePath } from './utils'
import { DateTime as NeoDateTime } from 'neo4j-driver/lib/v1/temporal-types.js'

async function addExifTags({ session, photo }) {
  const exifResult = await session.run(
    'MATCH (p:Photo { id: {id} })-[:EXIF]->(exif:PhotoEXIF) RETURN exif',
    {
      id: photo.id,
    }
  )

  if (exifResult.records.length > 0) return

  const rawTags = await exiftool.read(photo.path)

  const photoExif = {
    camera: rawTags.Model,
    maker: rawTags.Make,
    lens: rawTags.LensType,
    dateShot:
      rawTags.DateTimeOriginal &&
      NeoDateTime.fromStandardDate(rawTags.DateTimeOriginal.toDate()),
    fileSize: rawTags.FileSize,
    exposure: rawTags.ShutterSpeedValue,
    aperture: rawTags.ApertureValue,
    iso: rawTags.ISO,
    focalLength: rawTags.FocalLength,
    flash: rawTags.Flash,
  }

  const result = await session.run(
    `MATCH (p:Photo { id: {id} })
      CREATE (p)-[:EXIF]->(exif:PhotoEXIF {exifProps})`,
    {
      id: photo.id,
      exifProps: photoExif,
    }
  )

  console.log('Added exif tags to photo', photo.path)
}

export default async function processImage({ driver, markFinishedImage }, id) {
  const session = driver.session()

  const result = await session.run(
    `MATCH (p:Photo { id: {id} })<-[:CONTAINS]-(a:Album) RETURN p, a.id`,
    {
      id,
    }
  )

  const photo = result.records[0].get('p').properties
  const albumId = result.records[0].get('a.id')

  const imagePath = getImageCachePath(id, albumId)

  // Verify that processing is needed
  if (await fs.exists(path.resolve(imagePath, 'thumbnail.jpg'))) {
    const urlResult = await session.run(
      `MATCH (p:Photo { id: {id} })-->(urls:PhotoURL) RETURN urls`,
      { id }
    )

    if (urlResult.records.length == 2) {
      markFinishedImage()

      console.log('Skipping image', photo.path)
      return
    }
  }

  // Begin processing
  await session.run(
    `MATCH (p:Photo { id: {id} })-->(urls:PhotoURL) DETACH DELETE urls`,
    { id }
  )

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

  await addExifTags({ session, photo })

  session.close()

  markFinishedImage()
}
