import {
  InMemoryCache,
  ApolloClient,
  split,
  ApolloLink,
  HttpLink,
  ServerError,
  FieldMergeFunction,
} from '@apollo/client'
import { getMainDefinition } from '@apollo/client/utilities'
import { onError } from '@apollo/client/link/error'
import { WebSocketLink } from '@apollo/client/link/ws'

import urlJoin from 'url-join'
import { clearTokenCookie } from './helpers/authentication'
import { MessageState } from './components/messages/Messages'
import { Message } from './components/messages/SubscriptionsHook'
import { NotificationType } from './__generated__/globalTypes'

export const API_ENDPOINT = import.meta.env.REACT_APP_API_ENDPOINT
  ? (import.meta.env.REACT_APP_API_ENDPOINT as string)
  : urlJoin(location.origin, '/api')

export const GRAPHQL_ENDPOINT = urlJoin(API_ENDPOINT, '/graphql')

const httpLink = new HttpLink({
  uri: GRAPHQL_ENDPOINT,
  credentials: 'include',
})

console.log('GRAPHQL ENDPOINT', GRAPHQL_ENDPOINT)

const apiProtocol = new URL(GRAPHQL_ENDPOINT).protocol

const websocketUri = new URL(GRAPHQL_ENDPOINT)
websocketUri.protocol = apiProtocol === 'https:' ? 'wss:' : 'ws:'

const wsLink = new WebSocketLink({
  uri: websocketUri.toString(),
  // credentials: 'include',
})

const link = split(
  // split based on operation type
  ({ query }) => {
    const definition = getMainDefinition(query)
    return (
      definition.kind === 'OperationDefinition' &&
      definition.operation === 'subscription'
    )
  },
  wsLink,
  httpLink
)

const linkError = onError(({ graphQLErrors, networkError }) => {
  const errorMessages = []

  const formatPath = (path: readonly (string | number)[] | undefined) =>
    path?.join('::') ?? 'undefined'

  if (graphQLErrors) {
    graphQLErrors.map(({ message, locations, path }) =>
      console.log(
        `[GraphQL error]: Message: ${message}, Location: ${JSON.stringify(
          locations
        )} Path: ${formatPath(path)}`
      )
    )

    if (graphQLErrors.length == 1) {
      errorMessages.push({
        header: 'Something went wrong',
        content: `Server error: ${graphQLErrors[0].message} at (${formatPath(
          graphQLErrors[0].path
        )})`,
      })
    } else if (graphQLErrors.length > 1) {
      errorMessages.push({
        header: 'Multiple things went wrong',
        content: `Received ${graphQLErrors.length} errors from the server. See the console for more information`,
      })
    }

    if (graphQLErrors.find(x => x.message == 'unauthorized')) {
      console.log('Unauthorized, clearing token cookie')
      clearTokenCookie()
      // location.reload()
    }
  }

  if (networkError) {
    console.log(`[Network error]: ${JSON.stringify(networkError)}`)
    clearTokenCookie()

    const errors =
      ((networkError as ServerError)?.result.errors as Error[]) || []

    if (errors.length == 1) {
      errorMessages.push({
        header: 'Server error',
        content: `You are being logged out in an attempt to recover.\n${errors[0].message}`,
      })
    } else if (errors.length > 1) {
      errorMessages.push({
        header: 'Multiple server errors',
        content: `Received ${
          graphQLErrors?.length || 0
        } errors from the server. You are being logged out in an attempt to recover.`,
      })
    }
  }

  if (errorMessages.length > 0) {
    const newMessages: Message[] = errorMessages.map(msg => ({
      key: Math.random().toString(26),
      type: NotificationType.Message,
      props: {
        negative: true,
        ...msg,
      },
    }))
    MessageState.set((messages: Message[]) => [...messages, ...newMessages])
  }
})

type PaginateCacheType = {
  keyArgs: string[]
  merge: FieldMergeFunction<unknown[], unknown[]>
}

// Modified version of Apollo's offsetLimitPagination()
const paginateCache = (keyArgs: string[]) =>
  ({
    keyArgs,
    merge(existing, incoming, { args, fieldName }) {
      const merged = existing ? existing.slice(0) : []
      if (args?.paginate) {
        const { offset = 0 } = args.paginate as { offset: number }
        for (let i = 0; i < incoming.length; ++i) {
          merged[offset + i] = incoming[i]
        }
      } else {
        throw new Error(`Paginate argument is missing for query: ${fieldName}`)
      }
      return merged
    },
  } as PaginateCacheType)

const memoryCache = new InMemoryCache({
  typePolicies: {
    // There only exists one global instance of SiteInfo,
    // therefore it can always be merged
    SiteInfo: {
      merge: true,
    },
    MediaURL: {
      keyFields: ['url'],
    },
    Album: {
      fields: {
        media: paginateCache(['onlyFavorites', 'order']),
      },
    },
    FaceGroup: {
      fields: {
        imageFaces: paginateCache([]),
      },
    },
    Query: {
      fields: {
        myTimeline: paginateCache(['onlyFavorites']),
        myFaceGroups: paginateCache([]),
      },
    },
  },
})

const client = new ApolloClient({
  // link: ApolloLink.from([linkError, authLink.concat(link)]),
  link: ApolloLink.from([linkError, link]),
  cache: memoryCache,
})

export default client
