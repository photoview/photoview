import fs from 'fs-extra'
import path from 'path'
import { resolve as pathResolve, basename as pathBasename } from 'path'
import { PubSub } from 'apollo-server'
import uuid from 'uuid'
import { exiftool } from 'exiftool-vendored'
import sharp from 'sharp'
import { isImage, isRawImage } from './utils'
import { promisify } from 'util'
import config from '../config'
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

    this.addImageToProgress = () => {
      this.imagesToProgress++
    }

    this.addFinishedImage = () => {
      this.finishedImages++
    }
  }

  async scanUser(user) {
    await _scanUser({ driver: this.driver, scanAlbum: this.scanAlbum }, user)
  }

  async scanAlbum(album) {
    await _scanAlbum(
      {
        driver: this.driver,
        addImageToProgress: this.addImageToProgress,
        addFinishedImage: this.addFinishedImage,
        processImage: this.processImage,
      },
      album
    )

    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: this.finishedImages / this.imagesToProgress,
        finished: false,
        error: false,
        errorMessage: '',
      },
    })
  }

  async processImage(id) {
    await _processImage(
      {
        driver: this.driver,
        addFinishedImage: this.addFinishedImage,
      },
      id
    )
  }

  async scanAll() {
    this.pubsub.publish(EVENT_SCANNER_PROGRESS, {
      scannerStatusUpdate: {
        progress: 0,
        finished: false,
        error: false,
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
          error: true,
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
        error: false,
        errorMessage: '',
      },
    })
  }
}

export default PhotoScanner
