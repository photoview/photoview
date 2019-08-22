import fs from 'fs-extra'
import path from 'path'
import generateID from '../id-generator'
import { isImage, getImageCachePath } from './utils'
import _processImage from './processImage'
import { EVENT_SCANNER_PROGRESS } from './Scanner'

export default async function scanAlbum(scanner, album) {
  const { driver, markImageToProgress } = scanner
  const { title, path: albumPath, id } = album

  console.log('Scanning album', title)

  let processedImages = []

  const list = await fs.readdir(albumPath)
  let processingImagePromises = []

  const addPhotoToProcess = photo => {
    markImageToProgress(photo.id)

    processingImagePromises.push(
      _processImage(scanner, photo.id).catch(e => {
        console.error(`Error processing image (${e.path}): ${e.stack}`)
        scanner.pubsub.publish(EVENT_SCANNER_PROGRESS, {
          scannerStatusUpdate: {
            progress: 0,
            finished: false,
            success: false,
            message: `Error processing image at ${e.path}: ${e.message}`,
          },
        })
      })
    )
  }

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

        const photo = photoResult.records[0].get('p').properties

        addPhotoToProcess(photo)
      } else {
        console.log(`Found new image at ${itemPath}`)

        const photo = {
          id: generateID(),
          path: itemPath,
          title: item,
        }

        await session.run(
          `MATCH (a:Album { id: {albumId} })
          CREATE (p:Photo {photo})
          CREATE (a)-[:CONTAINS]->(p)`,
          {
            photo,
            albumId: id,
          }
        )

        addPhotoToProcess(photo)
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

  await Promise.all(processingImagePromises).catch(e => {
    console.error(`Error processing image: ${e.stack}`)
    scanner.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: false,
        success: false,
        message: `Error processing image: ${e.message}`,
      },
    })
  })
  console.log('Done processing album', album.title)

  scanner.broadcastProgress()
}
