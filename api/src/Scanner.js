import fs from 'fs-extra'
import path from 'path'
import { resolve as pathResolve, basename as pathBasename } from 'path'
import { PubSub } from 'apollo-server'
import uuid from 'uuid'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import readChunk from 'read-chunk'
import imageType from 'image-type'
import { promisify } from 'util'
import config from './config'

const imageSize = promisify(require('image-size'))

export const EVENT_SCANNER_PROGRESS = 'SCANNER_PROGRESS'

const isImage = async path => {
  const buffer = await readChunk(path, 0, 12)
  const type = imageType(buffer)

  return type != null
}

export const isRawImage = async path => {
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

    this.imagesToProgress = 0
    this.finishedImages = 0
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

    let allUserScans = []

    session.run('MATCH (u:User) return u').subscribe({
      onNext: record => {
        const user = record.toObject().u.properties

        console.log('USER', user)

        if (!user.rootPath) {
          console.log(`User ${user.username}, has no root path, skipping`)
          return
        }

        console.log(`Scanning ${user.username}...`)
        allUserScans.push(this.scanUser(user))
      },
      onCompleted: () => {
        session.close()
        this.isRunning = false

        Promise.all(allUserScans)
          .then(() => {
            console.log(
              `Done scanning ${this.finishedImages} of ${this.imagesToProgress}`
            )
            this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
              scannerStatusUpdate: {
                progress: 100,
                finished: true,
                error: false,
                errorMessage: '',
              },
            })
          })
          .catch(error => {
            console.log('SYNC ERROR', JSON.stringify(error))
            this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
              scannerStatusUpdate: {
                progress: 0,
                finished: false,
                error: true,
                errorMessage: error.message,
              },
            })
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

    let foundAlbumIds = []

    async function scanPath(path, parentAlbum) {
      const list = fs.readdirSync(path)

      let foundImage = false
      let newAlbums = []

      for (const item of list) {
        const itemPath = pathResolve(path, item)
        // console.log(`Scanning item ${itemPath}...`)
        const stat = fs.statSync(itemPath)

        if (stat.isDirectory()) {
          const session = driver.session()
          let nextParentAlbum = null

          const findAlbumResult = await session.run(
            'MATCH (a:Album { path: {path} }) RETURN a',
            {
              path: itemPath,
            }
          )

          session.close()

          if (findAlbumResult.records.length != 0) {
            const album = findAlbumResult.records[0].toObject().a.properties
            console.log('Found existing album', album.title)

            foundAlbumIds.push(album.id)

            nextParentAlbum = album.id
            scanAlbum(album)

            continue
          }

          const {
            foundImage: imagesInDirectory,
            newAlbums: childAlbums,
          } = await scanPath(itemPath, nextParentAlbum)

          if (imagesInDirectory) {
            console.log(`Found new album at ${itemPath}`)
            const session = driver.session()

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

            foundAlbumIds.push(albumId)
            newAlbums.push(albumId)
            const album = albumResult.records[0].toObject().a.properties

            if (parentAlbum) {
              console.log('Linking parent album for', album.title)
              await session.run(
                `MATCH (parent:Album { id: {parentId} })
                MATCH (child:Album { id: {childId} })
                CREATE (parent)-[:SUBALBUM]->(child)`,
                {
                  childId: albumId,
                  parentId: parentAlbum,
                }
              )
            }

            console.log(`Linking ${childAlbums.length} child albums`)
            for (let childAlbum of childAlbums) {
              await session.run(
                `MATCH (parent:Album { id: {parentId} })
                MATCH (child:Album { id: {childId} })
                CREATE (parent)-[:SUBALBUM]->(child)`,
                {
                  parentId: albumId,
                  childId: childAlbum,
                }
              )
            }

            scanAlbum(album)

            session.close()
          }

          continue
        }

        if (!foundImage && (await isImage(itemPath))) {
          foundImage = true
        }
      }

      return { foundImage, newAlbums }
    }

    await scanPath(user.rootPath)

    console.log('Found album ids', foundAlbumIds)

    const session = this.driver.session()

    const userAlbumsResult = await session.run(
      'MATCH (u:User { id: {userId} })-[:OWNS]->(a:Album)-[:CONTAINS]->(p:Photo) WHERE NOT a.id IN {foundAlbums} DETACH DELETE a, p RETURN a',
      { userId: user.id, foundAlbums: foundAlbumIds }
    )

    console.log(
      `Deleted ${userAlbumsResult.records.length} albums from ${user.username} that was not found locally`
    )

    session.close()

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

        this.imagesToProgress++

        const photoResult = await session.run(
          `MATCH (p:Photo {path: {imgPath} })<--(a:Album {id: {albumId}}) RETURN p`,
          {
            imgPath: itemPath,
            albumId: id,
          }
        )

        if (photoResult.records.length != 0) {
          console.log(`Photo already exists ${item}`)

          const id = photoResult.records[0].get('p').properties.id

          const thumbnailPath = pathResolve(
            config.cachePath,
            id,
            'thumbnail.jpg'
          )

          if (!(await fs.exists(thumbnailPath))) {
            this.processImage(id)
          } else {
            this.finishedImages++
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

          this.processImage(imageId)
        }
      }
    }
  }

  async processImage(id) {
    const session = this.driver.session()

    const result = await session.run(`MATCH (p:Photo { id: {id} }) RETURN p`, {
      id,
    })

    await session.run(
      `MATCH (p:Photo { id: {id} })-[rel]->(url:PhotoURL) DELETE url, rel`,
      { id }
    )

    console.log('PROCESS IMAGE RESULT', result)

    const photo = result.records[0].get('p').properties

    console.log('Processing photo', photo.path)

    const imagePath = path.resolve(config.cachePath, 'images', id)

    await fs.remove(imagePath)
    await fs.mkdirp(imagePath)

    let originalPath = photo.path

    if (await isRawImage(photo.path)) {
      console.log('Processing RAW image')

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

    session.close()

    console.log('Processing done')
    this.finishedImages++

    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: this.finishedImages / this.imagesToProgress,
        finished: false,
        error: false,
        errorMessage: '',
      },
    })
  }
}

export default PhotoScanner
