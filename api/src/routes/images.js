import fs from 'fs-extra'
import path from 'path'
import _ from 'lodash'
import config from '../config'
import { isRawImage, getImageCachePath } from '../scanner/utils'
import { getUserFromToken, getTokenFromBearer } from '../token'

async function sendImage({ imagePath, photo, res }) {
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

function loadImageRoutes({ app, driver, scanner }) {
  app.use('/images/:id/:image', async (req, res) => {
    const { id, image } = req.params

    let user = null

    try {
      const token = getTokenFromBearer(req.headers.authorization)
      user = await getUserFromToken(token, driver)
    } catch (err) {
      return res.status(401).send(err.message)
    }

    const session = driver.session()

    const result = await session.run(
      'MATCH (p:Photo { id: {id} })<-[:CONTAINS]-(a:Album)<-[:OWNS]-(u:User) RETURN p as photo, u.id as userId, a.id as albumId',
      {
        id,
      }
    )

    if (result.records.length == 0) {
      return res.status(404).send(`Image not found`)
    }

    const userId = result.records[0].get('userId')
    const albumId = result.records[0].get('albumId')
    const photo = result.records[0].get('photo').properties

    if (userId != user.id) {
      return res.status(401).send(`Image not owned by you`)
    }

    session.close()

    let imagePath = path.resolve(getImageCachePath(id, albumId), image)

    sendImage({ imagePath, photo, res })
  })

  // app.use('/share/:token/:image', async (req, res) => {
  //   const { token } = req.params
  // })
}

export default loadImageRoutes
