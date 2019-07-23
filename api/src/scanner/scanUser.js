import fs from 'fs-extra'
import { resolve as pathResolve } from 'path'
import uuid from 'uuid'
import { isImage, getAlbumCachePath } from './utils'

export default async function scanUser({ driver, scanAlbum }, user) {
  console.log('Scanning user', user.username, 'at', user.path)

  let foundAlbumIds = []
  let albumScanPromises = []

  async function scanPath(path, parentAlbum) {
    console.log('SCAN PATH', path)
    const list = fs.readdirSync(path)

    let foundImageOrAlbum = false
    let newAlbums = []

    for (const item of list) {
      const itemPath = pathResolve(path, item)
      // console.log(`Scanning item ${itemPath}...`)
      let stat = null

      try {
        stat = await fs.stat(itemPath)
      } catch {
        console.log('ERROR reading file stat for item:', itemPath)
      }

      if (stat && stat.isDirectory()) {
        const session = driver.session()
        let nextParentAlbum = null

        const findAlbumResult = await session.run(
          'MATCH (a:Album { path: {path} }) RETURN a',
          {
            path: itemPath,
          }
        )

        session.close()

        const {
          foundImage: imagesInDirectory,
          newAlbums: childAlbums,
        } = await scanPath(itemPath, nextParentAlbum)

        if (findAlbumResult.records.length > 0) {
          const album = findAlbumResult.records[0].toObject().a.properties
          console.log('Found existing album', album.title)

          foundImageOrAlbum = true
          nextParentAlbum = album.id
          foundAlbumIds.push(album.id)
          albumScanPromises.push(scanAlbum(album))

          continue
        }

        if (imagesInDirectory) {
          console.log(`Found new album at ${itemPath}`)
          foundImageOrAlbum = true

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

          newAlbums.push(albumId)
          const album = albumResult.records[0].toObject().a.properties

          if (parentAlbum) {
            console.log('Linking parent album for', album.title)
            await session.run(
              `MATCH (parent:Album { id: {parentId} })
              MATCH (child:Album { id: {childId} })
              MERGE (parent)-[:SUBALBUM]->(child)`,
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

          foundAlbumIds.push(album.id)
          albumScanPromises.push(scanAlbum(album))

          session.close()
        }

        continue
      }

      if (!foundImageOrAlbum && (await isImage(itemPath))) {
        foundImageOrAlbum = true
      }
    }

    return { foundImage: foundImageOrAlbum, newAlbums }
  }

  await scanPath(user.rootPath)

  const session = driver.session()

  const userAlbumsResult = await session.run(
    `MATCH (u:User { id: {userId} })-[:OWNS]->(a:Album)
    WHERE NOT a.id IN {foundAlbums}
    OPTIONAL MATCH (a)-[:CONTAINS]->(p:Photo)-->(photoTail)
    WITH a, p, photoTail, a.id AS albumId
    DETACH DELETE a, p, photoTail
    RETURN albumId`,
    { userId: user.id, foundAlbums: foundAlbumIds }
  )

  console.log('FOUND ALBUM IDS', foundAlbumIds)
  const deletedAlbumIds = userAlbumsResult.records.map(record =>
    record.get('albumId')
  )

  for (const albumId of deletedAlbumIds) {
    await fs.remove(getAlbumCachePath(albumId))
  }

  console.log(
    `Deleted ${userAlbumsResult.records.length} albums from ${user.username} that was not found locally`
  )

  session.close()

  await Promise.all(albumScanPromises)

  console.log('User scan complete')
}
