import { ApolloClient } from 'apollo-client'
import { InMemoryCache } from 'apollo-cache-inmemory'
import { HttpLink } from 'apollo-link-http'
import { WebSocketLink } from 'apollo-link-ws'
import { onError } from 'apollo-link-error'
import { setContext } from 'apollo-link-context'
import { ApolloLink, split } from 'apollo-link'
import { getMainDefinition } from 'apollo-utilities'

const httpLink = new HttpLink({
  uri: process.env.GRAPHQL_ENDPOINT,
  credentials: 'same-origin',
})

console.log('GRAPHQL ENDPOINT', process.env.GRAPHQL_ENDPOINT)

const apiProtocol = new URL(process.env.GRAPHQL_ENDPOINT).protocol

let websocketUri = new URL(process.env.GRAPHQL_ENDPOINT)
websocketUri.protocol = apiProtocol === 'https:' ? 'wss:' : 'ws:'

const wsLink = new WebSocketLink({
  uri: websocketUri,
  credentials: 'same-origin',
  options: {
    reconnect: true,
    connectionParams: {
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
  },
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
  if (graphQLErrors)
    graphQLErrors.map(({ message, locations, path }) =>
      console.log(
        `[GraphQL error]: Message: ${message}, Location: ${JSON.stringify(
          locations
        )} Path: ${path}`
      )
    )
  if (networkError) {
    console.log(`[Network error]: ${JSON.stringify(networkError)}`)
    localStorage.removeItem('token')
  }
})

const authLink = setContext((_, { headers }) => {
  // get the authentication token from local storage if it exists
  const token = localStorage.getItem('token')
  // return the headers to the context so httpLink can read them
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
    },
  }
})

const client = new ApolloClient({
  link: ApolloLink.from([linkError, authLink.concat(link)]),
  cache: new InMemoryCache(),
})

export default client
