import { getImageFromRequest, RequestError } from './images'
import fs from 'fs-extra'
import path from 'path'

async function sendDownload(req, res) {
  let { photo, cachePath } = req
  const cacheBasename = path.basename(cachePath)
  const photoBasename = path.basename(photo.path)

  if (cacheBasename == photoBasename) {
    if (!(await fs.exists(photo.path))) {
      throw new RequestError(500, 'Image missing from the server')
    }

    return res.sendFile(photo.path)
  }

  throw new RequestError(404, 'Image could not be found')
}

const loadDownloadRoutes = router => {
  router.use('/download/:id/:image', getImageFromRequest)
  router.use('/download/:id/:image', sendDownload)
}

export default loadDownloadRoutes
