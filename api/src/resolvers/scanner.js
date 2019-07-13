import { EVENT_SCANNER_PROGRESS } from '../scanner'

const mutation = {
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

const subscription = {
  scannerStatusUpdate: {
    subscribe(root, args, ctx, info) {
      return ctx.scanner.pubsub.asyncIterator([EVENT_SCANNER_PROGRESS])
    },
  },
}

export default {
  mutation,
  subscription,
}
