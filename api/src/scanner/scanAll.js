export default function scanAll({ driver, scanUser }) {
  return new Promise((resolve, reject) => {
    let session = driver.session()

    let allUserScans = []

    session.run('MATCH (u:User) return u').subscribe({
      onNext: record => {
        const user = record.toObject().u.properties

        if (!user.rootPath) {
          console.log(`User ${user.username}, has no root path, skipping`)
          return
        }

        allUserScans.push(
          scanUser(user).catch(reason => {
            console.log(
              `User scan exception for user ${user.username} ${reason}`
            )
            reject(reason)
          })
        )
      },
      onCompleted: () => {
        session.close()

        Promise.all(allUserScans).then(() => {
          resolve()
        })
      },
      onError: error => {
        session.close()
        reject(error)
      },
    })
  })
}
