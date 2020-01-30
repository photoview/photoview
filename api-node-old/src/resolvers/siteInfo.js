import { neo4jgraphql } from 'neo4j-graphql-js'
import jwt from 'jsonwebtoken'
import { registerUser, authorizeUser } from './users'

async function initialSetup(driver) {
  const session = driver.session()

  await session.run(
    `MERGE (info:SiteInfo) ON CREATE SET info = {initialSettings}`,
    {
      initialSettings: {
        initialSetup: true,
      },
    }
  )

  session.close()
}

const Query = {
  async siteInfo(root, args, ctx, info) {
    await initialSetup(ctx.driver)

    return neo4jgraphql(root, args, ctx, info)
  },
}

const Mutation = {
  async initialSetupWizard(root, args, ctx, info) {
    await initialSetup(ctx.driver)

    const session = ctx.driver.session()

    const result = await session.run(`MATCH (i:SiteInfo) RETURN i`)

    const siteInfo = result.records[0].get('i').properties

    if (siteInfo.initialSetup == false) {
      return {
        success: false,
        status: 'Has already been setup',
        token: null,
      }
    }

    const userResult = await registerUser(root, args, ctx, info)

    if (!userResult.success) {
      return userResult
    }

    const userId = jwt.decode(userResult.token).id

    await session.run(`MATCH (u:User { id: {id} }) SET u.admin = true`, {
      id: userId,
    })

    await session.run(`MATCH (i:SiteInfo) SET i.initialSetup = false`)

    session.close()

    const token = (await authorizeUser(root, args, ctx, info)).token

    return {
      success: true,
      status: 'Initial setup successful',
      token,
    }
  },
}

export default {
  Query,
  Mutation,
}
