import fs from 'fs-extra'
import path from 'path'
import { resolve as pathResolve, basename as pathBasename } from 'path'
import { PubSub } from 'apollo-server'
import uuid from 'uuid'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import readChunk from 'read-chunk'
import imageType from 'image-type'
import config from './config'

export const EVENT_SCANNER_PROGRESS = 'SCANNER_PROGRESS'

const isImage = async path => {
  const buffer = await readChunk(path, 0, 12)
  const type = imageType(buffer)

  return type != null
}

const isRawImage = async path => {
  const buffer = await readChunk(path, 0, 12)
  const { ext } = imageType(buffer)

  const rawTypes = ['cr2', 'arw', 'crw', 'dng']

  return rawTypes.includes(ext)
}

class PhotoScanner {
  constructor(driver) {
    this.driver = driver
    this.isRunning = false
    this.pubsub = new PubSub()

    this.scanAll = this.scanAll.bind(this)
    this.scanAlbum = this.scanAlbum.bind(this)
    this.scanUser = this.scanUser.bind(this)
    this.processImage = this.processImage.bind(this)
  }

  async scanAll() {
    this.isRunning = true

    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: false,
        error: false,
        errorMessage: '',
      },
    })

    let session = this.driver.session()

    session.run('MATCH (u:User) return u').subscribe({
      onNext: record => {
        const user = record.toObject().u.properties

        console.log('USER', user)

        if (!user.rootPath) {
          console.log(`User ${user.username}, has no root path, skipping`)
          return
        }

        this.scanUser(user)

        console.log(`Scanning ${user.username}...`)
      },
      onCompleted: () => {
        session.close()
        this.isRunning = false
        console.log('Done scanning')

        this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
          scannerStatusUpdate: {
            progress: 100,
            finished: true,
            error: false,
            errorMessage: '',
          },
        })
      },
      onError: error => {
        console.error(error)

        this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
          scannerStatusUpdate: {
            progress: 0,
            finished: false,
            error: true,
            errorMessage: error.message,
          },
        })
      },
    })
  }

  async scanUser(user) {
    console.log('Scanning path', user.rootPath)

    const driver = this.driver
    const scanAlbum = this.scanAlbum

    async function scanPath(path) {
      const list = fs.readdirSync(path)

      let foundImage = false

      for (const item of list) {
        const itemPath = pathResolve(path, item)
        // console.log(`Scanning item ${itemPath}...`)
        const stat = fs.statSync(itemPath)

        if (stat.isDirectory()) {
          // console.log(`Entering directory ${itemPath}`)
          const imagesInDirectory = await scanPath(itemPath)

          if (imagesInDirectory) {
            console.log(`Found album at ${itemPath}`)
            const session = driver.session()

            const findAlbumResult = await session.run(
              'MATCH (a:Album { path: {path} }) RETURN a',
              {
                path: itemPath,
              }
            )

            console.log('FIND ALBUM RESULT', findAlbumResult.records)

            if (findAlbumResult.records.length != 0) {
              console.log('Album already exists')

              const album = findAlbumResult.records[0].toObject().a.properties
              scanAlbum(album)

              continue
            }

            console.log('Adding album')
            const albumId = uuid()
            const albumResult = await session.run(
              `MATCH (u:User { id: {userId} })
              CREATE (a:Album { id: {id}, title: {title}, path: {path} })
              CREATE (u)-[own:OWNS]->(a)
              RETURN a`,
              {
                id: albumId,
                userId: user.id,
                title: item,
                path: itemPath,
              }
            )

            const album = albumResult.records[0].toObject().a.properties
            scanAlbum(album)

            session.close()
          }

          continue
        }

        if (!foundImage && (await isImage(itemPath))) {
          foundImage = true
        }
      }

      return foundImage
    }

    await scanPath(user.rootPath)
    console.log('User scan complete')
  }

  async scanAlbum(album) {
    const { title, path, id } = album
    console.log('Scanning album', title)

    const list = fs.readdirSync(path)

    for (const item of list) {
      const itemPath = pathResolve(path, item)

      if (await isImage(itemPath)) {
        const session = this.driver.session()

        const photoResult = await session.run(
          `MATCH (p:Photo {path: {imgPath} })<--(a:Album {id: {albumId}}) RETURN p`,
          {
            imgPath: itemPath,
            albumId: id,
          }
        )

        if (photoResult.records.length != 0) {
          console.log(`Photo already exists ${item}`)
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

          this.processImage(imageId)
        }
      }
    }
  }

  async processImage(id) {
    console.log('Processing image')
    const session = this.driver.session()

    const result = await session.run('MATCH (p:Photo { id: {id} }) return p', {
      id,
    })
    const photo = result.records[0].get('p').properties

    console.log('PHOTO', photo.path)

    const imagePath = path.resolve(config.cachePath, 'images', id)

    await fs.remove(imagePath)
    await fs.mkdirp(imagePath)

    let resizeBaseImg = photo.path

    if (await isRawImage(photo.path)) {
      console.log('Processing RAW image')

      const extractedPath = path.resolve(imagePath, 'extracted.jpg')
      await exiftool.extractPreview(photo.path, extractedPath)

      resizeBaseImg = extractedPath
    }

    // Resize image
    console.log('Resizing image', resizeBaseImg)
    await sharp(resizeBaseImg)
      .resize(1440, 1080, { fit: 'inside', withoutEnlargement: true })
      .toFile(path.resolve(imagePath, 'thumbnail.jpg'))

    await sharp(resizeBaseImg)
      .jpeg({ quality: 85 })
      .toFile(path.resolve(imagePath, 'original.jpg'))

    session.close()

    console.log('Processing done')
  }
}

export default PhotoScanner
