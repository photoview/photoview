import fs from 'fs-extra'
import path from 'path'
import generateID from '../id-generator'
import { isImage, getImageCachePath } from './utils'
import _processImage from './processImage'

export default async function scanAlbum(scanner, album) {
  const { driver, markImageToProgress } = scanner
  const { title, path: albumPath, id } = album

  console.log('Scanning album', title)

  let processedImages = []

  const list = await fs.readdir(albumPath)
  let processingImagePromises = []

  for (const item of list) {
    const itemPath = path.resolve(albumPath, item)
    processedImages.push(itemPath)

    if (await isImage(itemPath)) {
      const session = driver.session()

      const photoResult = await session.run(
        `MATCH (p:Photo {path: {imgPath} })<--(a:Album {id: {albumId}}) RETURN p`,
        {
          imgPath: itemPath,
          albumId: id,
        }
      )

      if (photoResult.records.length != 0) {
        // console.log(`Photo already exists ${item}`)

        const photoId = photoResult.records[0].get('p').properties.id
        markImageToProgress(photoId)

        const thumbnailPath = path.resolve(
          getImageCachePath(photoId, id),
          'thumbnail.jpg'
        )

        processingImagePromises.push(_processImage(scanner, photoId))
      } else {
        console.log(`Found new image at ${itemPath}`)

        const photoId = generateID()
        markImageToProgress(photoId)

        await session.run(
          `MATCH (a:Album { id: {albumId} })
          CREATE (p:Photo {id: {id}, path: {path}, title: {title} })
          CREATE (a)-[:CONTAINS]->(p)`,
          {
            id: photoId,
            path: itemPath,
            title: item,
            albumId: id,
          }
        )

        processingImagePromises.push(_processImage(scanner, photoId))
      }
    }
  }

  const session = driver.session()

  const deletedImagesResult = await session.run(
    `MATCH (a:Album { id: {albumId} })-[:CONTAINS]->(p:Photo)-->(trail)
    WHERE NOT p.path IN {images}
    WITH p, p.id AS imageId, trail
    DETACH DELETE p, trail
    RETURN DISTINCT imageId`,
    {
      albumId: id,
      images: processedImages,
    }
  )

  const deletedImages = deletedImagesResult.records.map(record =>
    record.get('imageId')
  )

  for (const imageId of deletedImages) {
    await fs.remove(getImageCachePath(imageId, id))
  }

  console.log(`Deleted ${deletedImages.length} images from album ${title}`)

  session.close()

  await Promise.all(processingImagePromises)
  console.log('Done processing album', album.title)

  scanner.broadcastProgress()
}
