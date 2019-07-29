import { neo4jgraphql } from 'neo4j-graphql-js'

async function initialSetup(driver) {
  const session = driver.session()

  await session.run(
    `MERGE (info:SiteInfo) ON CREATE SET info = {initialSettings}`,
    {
      initialSettings: {
        initialSetup: true,
        signupEnabled: false,
        defaultRoot: '/tmp',
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

    session.close()

    await session.run(
      `MERGE (info:SiteInfo) ON CREATE SET info = {initialSettings}`,
      {
        initialSettings: {
          initialSetup: true,
          signupEnabled: false,
        },
      }
    )

    session.close()

    return {
      success: true,
      status: 'Setup successful',
      token: null,
    }
  },
}

export default {
  Query,
  Mutation,
}
