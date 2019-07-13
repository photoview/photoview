import { PubSub } from 'apollo-server'

export const EVENT_SCANNER_PROGRESS = 'SCANNER_PROGRESS'

class PhotoScanner {
  constructor(driver) {
    this.driver = driver
    this.isRunning = false
    this.pubsub = new PubSub()
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

    session
      .run(
        'MATCH (u:User) return u.id AS id, u.username AS username, u.rootPath as path'
      )
      .subscribe({
        onNext: function(record) {
          const username = record.get('username')
          const id = record.get('id')
          const path = record.get('path')

          if (!path) {
            console.log(`User ${username}, has no root path, skipping`)
            return
          }

          this.scanUser(id)

          console.log(`Scanning ${username}...`)
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

  async scanUser(id) {}
}

export default PhotoScanner
