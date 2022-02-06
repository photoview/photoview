import { INITIAL_SETUP_QUERY } from './loginUtilities'

export const mockInitialSetupGraphql = (initial: boolean) => ({
  request: {
    query: INITIAL_SETUP_QUERY,
    variables: {},
  },
  result: {
    data: {
      siteInfo: {
        initialSetup: initial,
      },
    },
  },
})
