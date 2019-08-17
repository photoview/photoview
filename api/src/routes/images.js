import fs from 'fs-extra'
import path from 'path'
import _ from 'lodash'
import config from '../config'
import { isRawImage, getImageCachePath } from '../scanner/utils'
import { getUserFromToken, getTokenFromBearer } from '../token'

class RequestError extends Error {
  constructor(httpCode, message) {
    super(message)
    this.httpCode = httpCode
  }
}

async function sendImage({ photo, res, id, albumId, image, scanner }) {
  let imagePath = path.resolve(getImageCachePath(id, albumId), image)

  if (!(await fs.exists(imagePath))) {
    if (image == 'thumbnail.jpg') {
      console.log('Thumbnail not found, generating', photo.path)
      await scanner.processImage(photo.id)

      if (!(await fs.exists(imagePath))) {
        throw new Error('Thumbnail not found after image processing')
      }

      return res.sendFile(imagePath)
    }

    imagePath = photo.path
  }

  if (await isRawImage(imagePath)) {
    console.log('RAW preview image not found, generating', imagePath)
    await scanner.processImage(id)

    imagePath = path.resolve(config.cachePath, 'images', id, image)

    if (!(await fs.exists(imagePath))) {
      throw new Error('RAW preview not found after image processing')
    }

    return res.sendFile(imagePath)
  }

  res.sendFile(imagePath)
}

async function verifyUser({ req, driver, id }) {
  let user = null

  try {
    const token = getTokenFromBearer(req.headers.authorization)
    user = await getUserFromToken(token, driver)
  } catch (err) {
    throw new RequestError(401, err.message)
    // return res.status(401).send(err.message)
  }

  const session = driver.session()

  const result = await session.run(
    'MATCH (p:Photo { id: {id} })<-[:CONTAINS]-(a:Album)<-[:OWNS]-(u:User) RETURN p as photo, u.id as userId, a.id as albumId',
    {
      id,
    }
  )

  session.close()

  if (result.records.length == 0) {
    throw new RequestError(404, 'Image not found')
    // return res.status(404).send(`Image not found`)
  }

  const userId = result.records[0].get('userId')
  const albumId = result.records[0].get('albumId')
  const photo = result.records[0].get('photo').properties

  if (userId != user.id) {
    throw new RequestError(401, 'Image not owned by you')
    // return res.status(401).send(`Image not owned by you`)
  }

  return {
    user,
    albumId,
    photo,
  }
}

async function verifyShareToken({ shareToken, id, driver }) {
  const session = driver.session()

  const shareTokenResult = await session.run(
    `MATCH (share:ShareToken { token: {shareToken} })-[:SHARES]->(shared)
    MATCH (photo:Photo { id: {id} })<-[:CONTAINS]-(album:Album)
    RETURN share, photo, shared, album`,
    { shareToken, id }
  )

  session.close()

  if (shareTokenResult.records.length == 0) {
    throw new RequestError(404, 'Image not found')
  }

  const share = shareTokenResult.records[0].get('share').properties
  const album = shareTokenResult.records[0].get('album').properties
  const photo = shareTokenResult.records[0].get('photo').properties
  const sharedObject = shareTokenResult.records[0].get('shared')

  if (sharedObject.labels[0] == 'Album') {
    const session = driver.session()
    const albumResult = await session.run(
      `MATCH (album)-[:CONTAINS]->(photo:Photo { id: {id} })
      RETURN album`,
      { id }
    )
    session.close()

    if (albumResult.records.length == 0) {
      throw new RequestError(403, 'Invalid share token')
    }
  } else {
    const sharedPhoto = sharedObject.properties

    if (sharedPhoto.id != photo.id) {
      throw new RequestError(403, 'Invalid share token')
    }
  }

  return {
    photo,
    albumId: album.id,
  }
}

function loadImageRoutes({ app, driver, scanner }) {
  app.use('/images/:id/:image', async (req, res) => {
    const { id, image } = req.params

    const shareToken = req.query.token

    let photo, albumId

    try {
      let verify = null

      if (shareToken) {
        verify = await verifyShareToken({ shareToken, id, driver })
      }

      if (!verify) {
        verify = await verifyUser({
          req,
          driver,
          id,
        })
      }

      if (verify == null) throw RequestError(500, 'Unable to verify request')

      photo = verify.photo
      albumId = verify.albumId
    } catch (error) {
      return res.status(error.status || 500).send(error.message)
    }

    sendImage({ photo, res, id, albumId, image, scanner })
  })
}

export default loadImageRoutes
