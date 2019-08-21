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
