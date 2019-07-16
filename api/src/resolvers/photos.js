import { cypherQuery } from 'neo4j-graphql-js'
import { promisify } from 'util'
import fs from 'fs-extra'
import path from 'path'
import config from '../config'

const imageSize = promisify(require('image-size'))

function injectAt(query, index, injection) {
  return query.substr(0, index) + injection + query.substr(index)
}

const Query = {
  myAlbums: async function(root, args, ctx, info) {
    const query = cypherQuery(args, ctx, info)

    const whereSplit = query[0].indexOf('RETURN')

    query[0] = injectAt(
      query[0],
      whereSplit,
      `MATCH (u:User { id: {userid} }) WHERE (u)-[:OWNS]->(album) `
    )
    query[1].userid = ctx.user.id
    console.log(query)

    const addIDSplit = query[0].indexOf('album_photos {') + 14

    console.log('ID SPLIT', query[0].substr(0, addIDSplit))
    query[0] = injectAt(query[0], addIDSplit, `.id,`)

    const session = ctx.driver.session()

    const result = await session.run(...query)

    session.close()

    return result.records.map(record => record.get('album'))
  },
}

const Photo = {
  async original(root, args, ctx, info) {
    const imgPath = path.resolve(
      config.cachePath,
      'images',
      root.id,
      'original.jpg'
    )

    if (!(await fs.exists(imgPath))) {
      await ctx.scanner.processImage(root.id)
    }

    const { width, height } = await imageSize(imgPath)
    return {
      path: `${ctx.endpoint}/images/${root.id}/original.jpg`,
      width,
      height,
    }
  },
  async thumbnail(root, args, ctx, info) {
    return {
      path: `${ctx.endpoint}/images/${root.id}/thumbnail.jpg`,
      width: 120,
      height: 240,
    }
  },
}

export default {
  Query,
  Photo,
}
