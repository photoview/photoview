import { PubSub } from 'apollo-server'
import _ from 'lodash'
import _scanUser from './scanUser'
import _scanAlbum from './scanAlbum'
import _processImage from './processImage'
import _scanAll from './scanAll'

export const EVENT_SCANNER_PROGRESS = 'SCANNER_PROGRESS'

class PhotoScanner {
  constructor(driver) {
    this.driver = driver
    this.isRunning = false
    this.pubsub = new PubSub()

    this.processImage = this.processImage.bind(this)
    this.scanAlbum = this.scanAlbum.bind(this)
    this.scanUser = this.scanUser.bind(this)
    this.scanAll = this.scanAll.bind(this)

    this.imagesToProgress = 0
    this.finishedImages = 0

    this.markImageToProgress = () => {
      this.imagesToProgress++
    }

    this.markFinishedImage = () => {
      this.finishedImages++
    }

    this.broadcastProgress = _.debounce(() => {
      if (this.imagesToProgress == 0) return

      console.log(
        `Progress: ${(this.finishedImages / this.imagesToProgress) * 100}`
      )
      this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
        scannerStatusUpdate: {
          progress: (this.finishedImages / this.imagesToProgress) * 100,
          finished: false,
          success: true,
          errorMessage: '',
        },
      })
    }, 250)
  }

  async scanUser(user) {
    await _scanUser({ driver: this.driver, scanAlbum: this.scanAlbum }, user)
  }

  async scanAlbum(album) {
    await _scanAlbum(
      {
        driver: this.driver,
        markImageToProgress: this.markImageToProgress,
        markFinishedImage: this.markFinishedImage,
        processImage: this.processImage,
      },
      album
    )
  }

  async processImage(id) {
    await _processImage(
      {
        driver: this.driver,
        markFinishedImage: this.markFinishedImage,
      },
      id
    )

    this.broadcastProgress()
  }

  async scanAll() {
    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: false,
        success: true,
        errorMessage: '',
      },
    })

    try {
      await _scanAll({ driver: this.driver, scanUser: this.scanUser })
    } catch (error) {
      this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
        scannerStatusUpdate: {
          progress: 0,
          finished: false,
          success: false,
          errorMessage: error.message,
        },
      })
      throw error
    }

    console.log(
      `Done scanning ${this.finishedImages} of ${this.imagesToProgress}`
    )

    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 100,
        finished: true,
        success: true,
        errorMessage: '',
      },
    })
  }
}

export default PhotoScanner
