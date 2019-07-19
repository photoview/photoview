import { cypherQuery } from 'neo4j-graphql-js'
import { promisify } from 'util'
import fs from 'fs-extra'
import path from 'path'
import config from '../config'

const imageSize = promisify(require('image-size'))

function injectAt(query, index, injection) {
  return query.substr(0, index) + injection + query.substr(index)
}

const myAlbumQuery = function(args, ctx, info) {
  const query = cypherQuery(args, ctx, info)

  const whereSplit = query[0].indexOf('RETURN')

  query[0] = injectAt(
    query[0],
    whereSplit,
    `MATCH (u:User { id: {userid} }) WHERE (u)-[:OWNS]->(album) `
  )
  query[1].userid = ctx.user.id

  const addIDSplitIndex = query[0].indexOf('album_photos {')

  if (addIDSplitIndex != -1) {
    const addIDSplit = addIDSplitIndex + 14

    console.log('ID SPLIT', query[0])

    query[0] = injectAt(
      query[0],
      addIDSplit,
      query[0].indexOf('album_photos {}') == -1 ? ` .id, ` : ` .id `
    )
  }

  return query
}

const myPhotoQuery = function(args, ctx, info) {
  const query = cypherQuery(args, ctx, info)

  const whereSplit = query[0].indexOf('RETURN')

  query[0] = injectAt(
    query[0],
    whereSplit,
    `MATCH (u:User { id: {userid} }) WHERE (u)-[:OWNS]->(:Album)-[:CONTAINS]->(photo) `
  )
  query[1].userid = ctx.user.id

  query[0] = injectAt(
    query[0],
    query[0].indexOf('RETURN `photo` {') + 16,
    query[0].indexOf('RETURN `photo` {}') == -1 ? ` .id, ` : ` .id `
  )

  return query
}

const Query = {
  async myAlbums(root, args, ctx, info) {
    let query = myAlbumQuery(args, ctx, info)
    console.log(query)

    const session = ctx.driver.session()

    const result = await session.run(...query)

    session.close()

    return result.records.map(record => record.get('album'))
  },
  async album(root, args, ctx, info) {
    const session = ctx.driver.session()

    let query = myAlbumQuery(args, ctx, info)

    const whereSplit = query[0].indexOf('RETURN')

    query[0] = injectAt(query[0], whereSplit, ` AND album.id = {id} `)
    query[1].id = args.id
    console.log(query)

    const result = await session.run(...query)

    session.close()

    if (result.records.length == 0) {
      throw new Error('Album was not found')
    }

    return result.records[0].get('album')
  },
  async myPhotos(root, args, ctx, info) {
    let query = myPhotoQuery(args, ctx, info)
    console.log(query)

    const session = ctx.driver.session()

    const result = await session.run(...query)

    session.close()

    return result.records.map(record => record.get('photo'))
  },
  async photo(root, args, ctx, info) {
    const session = ctx.driver.session()

    let query = myPhotoQuery(args, ctx, info)

    const whereSplit = query[0].indexOf('RETURN')

    query[0] = injectAt(query[0], whereSplit, ` AND photo.id = {id} `)
    query[1].id = args.id
    console.log(query)

    const result = await session.run(...query)

    session.close()

    if (result.records.length == 0) {
      throw new Error('Album was not found')
    }

    return result.records[0].get('photo')
  },
}

function photoResolver(image) {
  return async (root, args, ctx, info) => {
    const imgPath = path.resolve(config.cachePath, 'images', root.id, image)

    if (!(await fs.exists(imgPath))) {
      await ctx.scanner.processImage(root.id)
    }

    const { width, height } = await imageSize(imgPath)
    return {
      path: `${ctx.endpoint}/images/${root.id}/${image}`,
      width,
      height,
    }
  }
}

const Photo = {
  original: photoResolver('original.jpg'),
  thumbnail: photoResolver('thumbnail.jpg'),
}

export default {
  Query,
  Photo,
}
