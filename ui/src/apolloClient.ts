import {
  InMemoryCache,
  ApolloClient,
  split,
  ApolloLink,
  HttpLink,
  ServerError,
  FieldMergeFunction,
  DocumentNode,
} from '@apollo/client'
import { getMainDefinition } from '@apollo/client/utilities'
import i18n from 'i18next'
import { onError } from '@apollo/client/link/error'
import { WebSocketLink } from '@apollo/client/link/ws'

import urlJoin from 'url-join'
import { authToken, clearTokenCookie } from './helpers/authentication'
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
  options: {
    reconnect: true,
    lazy: true,
    connectionParams: () => {
      const token = authToken()
      if (token) {
        return {
          Authorization: `Bearer ${token}`,
        }
      }
      return {}
    },
  },
})

const isSubscriptionOperation = (query: DocumentNode) => {
  const definition = getMainDefinition(query)
  return (
    definition.kind === 'OperationDefinition' &&
    definition.operation === 'subscription'
  )
}

const link = split(
  // split based on operation type
  ({ query }) => isSubscriptionOperation(query),
  wsLink,
  httpLink
)

// Exported for its tests; the client below is the only other user.
export const linkError = onError(
  ({ graphQLErrors, networkError, operation }) => {
    const errorMessages = []

    // The notification subscription reconnects on its own, e.g. after the
    // server restarts, and can transiently see a stale or missing token while
    // doing so. That says nothing about the user's own session, so it must not
    // log them out.
    const isSubscription = isSubscriptionOperation(operation.query)

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
          header: i18n.t(
            'notification.error.graphql.header',
            'Something went wrong'
          ),
          content: i18n.t(
            'notification.error.graphql.content',
            'Server error: {{message}} at ({{path}})',
            {
              message: graphQLErrors[0].message,
              path: formatPath(graphQLErrors[0].path),
            }
          ),
        })
      } else if (graphQLErrors.length > 1) {
        errorMessages.push({
          header: i18n.t(
            'notification.error.graphql.multiple_header',
            'Multiple things went wrong'
          ),
          content: i18n.t(
            'notification.error.graphql.multiple_content',
            'Received {{errorCount}} errors from the server. See the console for more information',
            { errorCount: graphQLErrors.length }
          ),
        })
      }
    }

    if (networkError) {
      console.log(`[Network error]: ${JSON.stringify(networkError)}`)

      // Only an actual authentication failure invalidates the token. A
      // timeout, an offline client or a server-side 500 would otherwise log
      // the user out over an outage that says nothing about their session.
      //
      // The status is the whole signal. The API turns an unknown or expired
      // token away with a 401 before any resolver runs. A GraphQL
      // `unauthorized` arrives with a 200 and means either that no token was
      // sent, so there is nothing to clear, or that this user may not touch
      // that object - sharing an album they do not own - which is no reason
      // to end their session.
      const statusCode = (networkError as ServerError | undefined)?.statusCode
      const loggingOut =
        !isSubscription && (statusCode === 401 || statusCode === 403)

      if (loggingOut) {
        clearTokenCookie()
      }

      // A plain connection failure has no result at all, and reading through
      // it used to throw from inside the error handler itself.
      const errors =
        ((networkError as ServerError)?.result?.errors as Error[]) || []

      const recoveryNote = loggingOut
        ? ` ${i18n.t(
            'notification.error.logout_note',
            'You are being logged out in an attempt to recover.'
          )}`
        : ''

      // Apollo hands a response's own `errors` to this handler twice: as
      // graphQLErrors, already reported above, and again inside
      // networkError.result. Reporting both showed one failure as two messages.
      const alreadyReported = (graphQLErrors?.length ?? 0) > 0

      if (alreadyReported) {
        if (loggingOut) {
          errorMessages.push({
            header: i18n.t(
              'notification.error.session.header',
              'Session ended'
            ),
            content: `${i18n.t(
              'notification.error.session.content',
              'The server no longer accepts this session.'
            )}${recoveryNote}`,
          })
        }
      } else if (errors.length == 1) {
        errorMessages.push({
          header: i18n.t('notification.error.server.header', 'Server error'),
          content: `${errors[0].message}${recoveryNote}`,
        })
      } else if (errors.length > 1) {
        errorMessages.push({
          header: i18n.t(
            'notification.error.server.multiple_header',
            'Multiple server errors'
          ),
          content: `${i18n.t(
            'notification.error.server.multiple_content',
            'Received {{errorCount}} errors from the server.',
            { errorCount: errors.length }
          )}${recoveryNote}`,
        })
      } else {
        // A connection that never reached the server carries no errors to
        // report, which would otherwise leave the user with nothing at all.
        errorMessages.push({
          header: i18n.t('notification.error.network.header', 'Network error'),
          content: `${i18n.t(
            'notification.error.network.content',
            'Could not reach the server.'
          )}${recoveryNote}`,
        })
      }
    }

    // A subscription reports its own errors: SubscriptionsHook shows them as
    // they arrive. Adding a message here as well showed every one twice - and
    // with reconnect enabled, a server that stays down repeats that on every
    // attempt. The log lines above still cover subscriptions.
    if (errorMessages.length > 0 && !isSubscription) {
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
  }
)

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
