import { EVENT_SCANNER_PROGRESS } from '../scanner/Scanner'

const Mutation = {
  async scanAll(root, args, ctx, info) {
    ctx.scanner.scanAll()

    return {
      finished: false,
      success: true,
      progress: 0,
      message: 'Starting scanner',
    }
  },
  async scanUser(root, args, ctx, info) {
    const session = ctx.driver.session()

    const userResult = await session.run(
      `MATCH (u:User { id: {userId} }) RETURN u`,
      {
        userId: args.userId,
      }
    )

    session.close()

    if (userResult.records.length == 0) {
      return {
        finished: false,
        success: false,
        progress: 0,
        message: 'Could not scan user: User not found',
      }
    }

    const user = userResult.records[0].get('u').properties

    ctx.scanner.scanUser(user)

    return {
      finished: false,
      success: true,
      progress: 0,
      message: 'Starting scanner',
    }
  },
}

const Subscription = {
  scannerStatusUpdate: {
    subscribe(root, args, ctx, info) {
      return ctx.scanner.pubsub.asyncIterator([EVENT_SCANNER_PROGRESS])
    },
  },
}

export default {
  Mutation,
  Subscription,
}
