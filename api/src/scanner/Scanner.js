import { PubSub } from 'apollo-server'
import _ from 'lodash'
import _scanUser from './scanUser'
import _scanAlbum from './scanAlbum'
import _processImage from './processImage'
import _scanAll from './scanAll'

export const EVENT_SCANNER_PROGRESS = 'SCANNER_PROGRESS'

async function _execScan(scanner, scanFunction) {
  try {
    if (scanner.isRunning) throw new Error('Scanner already running')
    scanner.isRunning = true
    scanner.imageProgress = {}

    const session = scanner.driver.session()
    const photoResult = await session.run(
      'MATCH (photo:Photo) RETURN photo.id as photoId'
    )
    session.close()

    photoResult.records
      .map(x => x.get('photoId'))
      .forEach(id => {
        scanner.markImageToProgress(id)
      })

    scanner.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: false,
        success: true,
        message: 'Scan started',
      },
    })

    console.log('Calling scan function')
    await scanFunction()
    console.log('Scan function ended')

    console.log(
      `Done scanning ${Object.keys(scanner.imageProgress).length} photos`
    )

    scanner.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 100,
        finished: true,
        success: true,
        message: `Done scanning ${
          Object.keys(scanner.imageProgress).length
        } photos`,
      },
    })
  } catch (e) {
    console.error(`SCANNER ERROR: ${e.message}\n${e.stack}`)
    scanner.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: true,
        success: false,
        message: `Scanner error: ${e.message}`,
      },
    })
  } finally {
    scanner.isRunning = false
  }
}

class PhotoScanner {
  constructor(driver) {
    this.driver = driver
    this.isRunning = false
    this.pubsub = new PubSub()

    this.processImage = this.processImage.bind(this)
    this.scanAlbum = this.scanAlbum.bind(this)
    this.scanUser = this.scanUser.bind(this)
    this.scanAll = this.scanAll.bind(this)

    this.imageProgress = {}

    this.markImageToProgress = imageId => {
      if (!this.imageProgress[imageId]) this.imageProgress[imageId] = false
    }

    this.markFinishedImage = imageId => {
      this.imageProgress[imageId] = true
      this.broadcastProgress()
    }

    this.finishedImages = () =>
      Object.values(this.imageProgress).reduce((prev, x) => {
        x ? prev++ : prev
        return prev
      }, 0)

    this.broadcastProgress = _.debounce(() => {
      if (!this.isRunning) return
      if (Object.keys(this.imageProgress).length == 0) return

      let progress =
        (this.finishedImages() / Object.keys(this.imageProgress).length) * 100

      console.log(`Progress: ${progress}`)
      this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
        scannerStatusUpdate: {
          progress,
          finished: false,
          success: true,
          message: `${this.finishedImages()} photos scanned`,
        },
      })
    }, 250)

    this.markImageToProgress = this.markImageToProgress.bind(this)
    this.markFinishedImage = this.markFinishedImage.bind(this)
    this.finishedImages = this.finishedImages.bind(this)
  }

  async scanUser(user) {
    await _execScan(this, async () => {
      await _scanUser({ driver: this.driver, scanAlbum: this.scanAlbum }, user)
    })
  }

  async scanAlbum(album) {
    await _execScan(this, async () => {
      await _scanAlbum(this, album)
    })
  }

  async processImage(id) {
    await _execScan(this, async () => {
      await _processImage(this, id)
    })
  }

  async scanAll() {
    await _execScan(this, async () => {
      await _scanAll(this)
    })
  }
}

export default PhotoScanner
