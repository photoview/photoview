import fs from 'fs-extra'
import path from 'path'
import uuid from 'uuid'
import { isImage } from './utils'
import config from '../config'

export default async function scanAlbum(
  { driver, addImageToProgress, addFinishedImage, processImage },
  album
) {
  const { title, path: albumPath, id } = album
  console.log('Scanning album', title)

  const list = await fs.readdir(albumPath)
  let processingImagePromises = []

  for (const item of list) {
    const itemPath = path.resolve(albumPath, item)

    if (await isImage(itemPath)) {
      const session = driver.session()

      addImageToProgress()

      const photoResult = await session.run(
        `MATCH (p:Photo {path: {imgPath} })<--(a:Album {id: {albumId}}) RETURN p`,
        {
          imgPath: itemPath,
          albumId: id,
        }
      )

      if (photoResult.records.length != 0) {
        // console.log(`Photo already exists ${item}`)

        const id = photoResult.records[0].get('p').properties.id

        const thumbnailPath = path.resolve(
          config.cachePath,
          id,
          'thumbnail.jpg'
        )

        if (!(await fs.exists(thumbnailPath))) {
          processingImagePromises.push(processImage(id))
        } else {
          addFinishedImage()
        }
      } else {
        console.log(`Found new image at ${itemPath}`)
        const imageId = uuid()
        await session.run(
          `MATCH (a:Album { id: {albumId} })
          CREATE (p:Photo {id: {id}, path: {path}, title: {title} })
          CREATE (a)-[:CONTAINS]->(p)`,
          {
            id: imageId,
            path: itemPath,
            title: item,
            albumId: id,
          }
        )

        processingImagePromises.push(processImage(imageId))
      }
    }
  }

  await Promise.all(processingImagePromises)
  console.log('Done processing album', album.title)
}
