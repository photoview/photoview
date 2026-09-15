import {
  ApolloLink,
  execute,
  FetchResult,
  gql,
  Observable,
} from '@apollo/client'
import { GraphQLError } from 'graphql'
import * as authentication from './helpers/authentication'
import { MessageState } from './components/messages/Messages'
import { linkError } from './apolloClient'

vi.mock('./helpers/authentication')

const QUERY = gql`
  query anyQuery {
    siteInfo {
      initialSetup
    }
  }
`

const SUBSCRIPTION = gql`
  subscription anySubscription {
    notification {
      key
    }
  }
`

/** An error as the HTTP link reports a response it could not use. */
const httpError = (statusCode: number, errors: { message: string }[] = []) =>
  Object.assign(new Error(`Response not successful: ${statusCode}`), {
    name: 'ServerError',
    statusCode,
    result: { errors },
  })

/** Runs one operation through the error link against a canned outcome. */
const run = (
  query: typeof QUERY,
  outcome: { error?: Error; result?: FetchResult }
) =>
  new Promise<void>(resolve => {
    const terminating = new ApolloLink(
      () =>
        new Observable(observer => {
          if (outcome.error) {
            observer.error(outcome.error)

            return
          }
          observer.next(outcome.result ?? { data: null })
          observer.complete()
        })
    )

    execute(ApolloLink.from([linkError, terminating]), { query }).subscribe({
      next: () => undefined,
      error: () => resolve(),
      complete: () => resolve(),
    })
  })

const clearTokenCookie = vi.mocked(authentication.clearTokenCookie)
let messagesAdded: number

beforeEach(() => {
  clearTokenCookie.mockClear()
  messagesAdded = 0
  vi.spyOn(MessageState, 'set').mockImplementation(update => {
    const added = (update as (messages: unknown[]) => unknown[])([])
    messagesAdded += added.length
  })
})

describe('what logs the user out', () => {
  test('an HTTP 401 clears the session', async () => {
    await run(QUERY, { error: httpError(401) })
    expect(clearTokenCookie).toHaveBeenCalled()
  })

  test('an HTTP 401 with an error body reports the error and the logout, once each', async () => {
    await run(QUERY, {
      error: httpError(401, [{ message: 'invalid authorization token' }]),
    })

    expect(clearTokenCookie).toHaveBeenCalled()
    expect(messagesAdded).toBe(2)
  })

  test('an HTTP 403 clears the session', async () => {
    await run(QUERY, { error: httpError(403) })
    expect(clearTokenCookie).toHaveBeenCalled()
  })

  test('a GraphQL unauthorized error clears the session', async () => {
    await run(QUERY, { result: { errors: [new GraphQLError('unauthorized')] } })
    expect(clearTokenCookie).toHaveBeenCalled()
  })
})

describe('what does not', () => {
  test('a server error keeps the session and says so once', async () => {
    await run(QUERY, { error: httpError(500, [{ message: 'boom' }]) })

    expect(clearTokenCookie).not.toHaveBeenCalled()
    expect(messagesAdded).toBe(1)
  })

  test('a connection that never reached the server keeps the session and does not throw', async () => {
    // No status, no result: reading the error list through it used to throw
    // from inside the handler itself.
    await run(QUERY, { error: new TypeError('Failed to fetch') })

    expect(clearTokenCookie).not.toHaveBeenCalled()
    expect(messagesAdded).toBe(1)
  })

  test('a failing subscription keeps the session and leaves the message to its own hook', async () => {
    // The notification subscription reports its errors itself. A second
    // message from here showed each one twice, again on every reconnect.
    await run(SUBSCRIPTION, { error: httpError(401) })
    await run(SUBSCRIPTION, { error: new TypeError('socket closed') })

    expect(clearTokenCookie).not.toHaveBeenCalled()
    expect(messagesAdded).toBe(0)
  })
})
