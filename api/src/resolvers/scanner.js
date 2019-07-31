import { EVENT_SCANNER_PROGRESS } from '../scanner/Scanner'

const Mutation = {
  async scanAll(root, args, ctx, info) {
    if (ctx.scanner.isRunning) {
      return {
        finished: false,
        success: false,
        errorMessage: 'Scanner already running',
      }
    }

    ctx.scanner.scanAll()

    return {
      finished: false,
      success: true,
      progress: 0,
      errorMessage: null,
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
