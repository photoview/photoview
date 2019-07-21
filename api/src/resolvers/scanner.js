import { EVENT_SCANNER_PROGRESS } from '../scanner/Scanner'

const Mutation = {
  async scanAll(root, args, ctx, info) {
    if (ctx.scanner.isRunning) {
      return {
        finished: false,
        error: true,
        errorMessage: 'Scanner already running',
      }
    }

    ctx.scanner.scanAll()

    return {
      finished: false,
      error: false,
      progress: 0,
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
