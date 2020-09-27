import { InMemoryCache } from 'apollo-cache-inmemory'
import { ApolloClient } from 'apollo-client'
import { ApolloLink, split } from 'apollo-link'
import { onError } from 'apollo-link-error'
import { HttpLink } from 'apollo-link-http'
import { WebSocketLink } from 'apollo-link-ws'
import { getMainDefinition } from 'apollo-utilities'
import urlJoin from 'url-join'
import { clearTokenCookie } from './authentication'
import { MessageState } from './components/messages/Messages'

export const GRAPHQL_ENDPOINT = process.env.API_ENDPOINT
  ? urlJoin(process.env.API_ENDPOINT, '/graphql')
  : urlJoin(location.origin, '/api/graphql')

const httpLink = new HttpLink({
  uri: GRAPHQL_ENDPOINT,
  credentials: 'include',
})

console.log('GRAPHQL ENDPOINT', GRAPHQL_ENDPOINT)

const apiProtocol = new URL(GRAPHQL_ENDPOINT).protocol

let websocketUri = new URL(GRAPHQL_ENDPOINT)
websocketUri.protocol = apiProtocol === 'https:' ? 'wss:' : 'ws:'

const wsLink = new WebSocketLink({
  uri: websocketUri,
  credentials: 'include',
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
  let errorMessages = []

  if (graphQLErrors) {
    graphQLErrors.map(({ message, locations, path }) =>
      console.log(
        `[GraphQL error]: Message: ${message}, Location: ${JSON.stringify(
          locations
        )} Path: ${path}`
      )
    )

    if (graphQLErrors.length == 1) {
      errorMessages.push({
        header: 'Something went wrong',
        content: `Server error: ${graphQLErrors[0].message} at (${graphQLErrors[0].path})`,
      })
    } else if (graphQLErrors.length > 1) {
      errorMessages.push({
        header: 'Multiple things went wrong',
        content: `Received ${graphQLErrors.length} errors from the server. See the console for more information`,
      })
    }
  }

  if (networkError) {
    console.log(`[Network error]: ${JSON.stringify(networkError)}`)
    clearTokenCookie()

    const errors = networkError.result.errors

    if (errors.length == 1) {
      errorMessages.push({
        header: 'Server error',
        content: `You are being logged out in an attempt to recover.\n${errors[0].message}`,
      })
    } else if (errors.length > 1) {
      errorMessages.push({
        header: 'Multiple server errors',
        content: `Received ${graphQLErrors.length} errors from the server. You are being logged out in an attempt to recover.`,
      })
    }
  }

  if (errorMessages.length > 0) {
    const newMessages = errorMessages.map(msg => ({
      key: Math.random().toString(26),
      type: 'message',
      props: {
        negative: true,
        ...msg,
      },
    }))
    MessageState.set(messages => [...messages, ...newMessages])
  }
})

const client = new ApolloClient({
  // link: ApolloLink.from([linkError, authLink.concat(link)]),
  link: ApolloLink.from([linkError, link]),
  cache: new InMemoryCache(),
})

export default client
