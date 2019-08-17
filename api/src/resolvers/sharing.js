import { neo4jgraphql } from 'neo4j-graphql-js'
import generateID from '../id-generator'
import { replaceMatch } from './neo4j-helpers'

const Mutation = {
  async shareAlbum(root, args, ctx, info) {
    const session = ctx.driver.session()

    const ownsAlbumResult = await session.run(
      `
      MATCH (u:User { id: {userId} })-[:OWNS]->(a:Album { id: {albumId} })
      RETURN a
    `,
      {
        userId: ctx.user.id,
        albumId: args.albumId,
      }
    )

    if (ownsAlbumResult.records.length == 0) {
      session.close()
      throw new Error('User does not own that album')
    }

    const createResult = await session.run(
      `
      MATCH (u:User { id: {userId} })
      MATCH (a:Album { id: {albumId} })
      CREATE (share:ShareToken {shareToken} )
      CREATE (u)-[:SHARE_TOKEN]->(share)-[:SHARES]->(a)
      RETURN share
    `,
      {
        userId: ctx.user.id,
        albumId: args.albumId,
        shareToken: {
          token: generateID(),
          expire: args.expire,
          password: args.password,
        },
      }
    )

    session.close()

    return {
      expire: null,
      password: null,
      ...createResult.records[0].get('share').properties,
    }
  },
  async sharePhoto(root, args, ctx, info) {
    const session = ctx.driver.session()

    const ownsPhotoResult = await session.run(
      `
      MATCH (u:User { id: {userId} })-[:OWNS]->(a:Album)-[:CONTAINS]->(p:Photo { id: {photoId} })
      RETURN a
    `,
      {
        userId: ctx.user.id,
        photoId: args.photoId,
      }
    )

    if (ownsPhotoResult.records.length == 0) {
      session.close()
      throw new Error('User does not own that photo')
    }

    const createResult = await session.run(
      `
      MATCH (u:User { id: {userId} })
      MATCH (p:Photo { id: {photoId} })
      CREATE (share:ShareToken {shareToken} )
      CREATE (u)-[:SHARE_TOKEN]->(share)-[:SHARES]->(p)
      RETURN share
    `,
      {
        userId: ctx.user.id,
        photoId: args.photoId,
        shareToken: {
          token: generateID(),
          expire: args.expire,
          password: args.password,
        },
      }
    )

    session.close()

    return {
      expire: null,
      password: null,
      ...createResult.records[0].get('share').properties,
    }
  },
  async deleteShareToken(root, args, ctx, info) {
    if (!ctx.user.admin) {
      const session = ctx.driver.session()
      const result = await session.run(
        `MATCH (u:User { id: {userId} })-[:SHARE_TOKEN]->(token:ShareToken { token: {token} })
         RETURN token`,
        {
          userId: ctx.user.id,
          token: args.token,
        }
      )

      session.close()

      if (result.records.length == 0) {
        throw new Error('User is not allowed to delete this share')
      }
    }

    return neo4jgraphql(root, args, ctx, info)
  },
}

const Query = {
  async albumShares(root, args, ctx, info) {
    const query = replaceMatch(
      { root, args, ctx, info },
      `
      MATCH (u:User { id: {userId} })
      MATCH (u)-[:OWNS]->(a:Album { id: {albumId} })
      MATCH (a)<-[:SHARES]-(shareToken:ShareToken)
    `
    )

    const session = ctx.driver.session()

    const queryResult = await session.run(query, {
      ...args,
      userId: ctx.user.id,
      albumId: args.id,
    })

    session.close()

    const tokens = queryResult.records.map(token => token.get('shareToken'))

    return tokens
  },
  async photoShares(root, args, ctx, info) {
    const query = replaceMatch(
      { root, args, ctx, info },
      `
      MATCH (u:User { id: {userId} })
      MATCH (u)-[:OWNS]->(a:Album)-[:CONTAINS]->(p:Photo {id: {photoId} })
      MATCH (p)<-[:SHARES]-(shareToken:ShareToken)
    `
    )

    const session = ctx.driver.session()

    const queryResult = await session.run(query, {
      ...args,
      userId: ctx.user.id,
      photoId: args.id,
    })

    session.close()

    const tokens = queryResult.records.map(token => token.get('shareToken'))

    return tokens
  },
  shareToken(root, args, ctx, info) {
    ctx.shareToken = args.token
    return neo4jgraphql(root, args, ctx, info)
  },
}

export default {
  Mutation,
  Query,
}
