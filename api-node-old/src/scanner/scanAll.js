import _scanUser from './scanUser'

export default function scanAll(scanner) {
  const { driver } = scanner
  return new Promise((resolve, reject) => {
    let session = driver.session()

    let usersToScan = []

    session.run('MATCH (u:User) return u').subscribe({
      onNext: record => {
        const user = record.toObject().u.properties

        if (!user.rootPath) {
          console.log(`User ${user.username}, has no root path, skipping`)
          return
        }

        usersToScan.push(user)
      },
      onCompleted: async () => {
        session.close()

        for (let user of usersToScan) {
          try {
            await _scanUser(scanner, user)
          } catch (reason) {
            console.log(
              `User scan exception for user ${user.username} ${reason}`
            )
            reject(reason)
          }
        }

        resolve()
      },
      onError: error => {
        session.close()
        reject(error)
      },
    })
  })
}
