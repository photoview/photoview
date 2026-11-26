import React from 'react'
import {
  ApolloClient,
  ApolloLink,
  ApolloProvider,
  InMemoryCache,
  Observable,
} from '@apollo/client'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ScannerJobStatus } from '../../__generated__/globalTypes'
import * as authentication from '../../helpers/authentication'
import SettingsPage from './SettingsPage'

vi.mock('../../helpers/authentication')

// Answers the admin check and the scanner queue. Everything else on the page
// stays pending, so the test does not depend on the other settings queries.
const clientFor = (isAdmin: boolean) =>
  new ApolloClient({
    cache: new InMemoryCache(),
    link: new ApolloLink(operation => {
      if (operation.operationName === 'adminQuery') {
        return Observable.of({
          data: { myUser: { __typename: 'User', admin: isAdmin } },
        })
      }

      if (operation.operationName === 'scannerQueueStatusQuery') {
        return Observable.of({
          data: {
            scannerQueueStatus: [
              {
                __typename: 'ScannerQueueItem',
                status: ScannerJobStatus.RUNNING,
                album: {
                  __typename: 'Album',
                  id: '7',
                  title: 'Holiday',
                  path: [],
                },
              },
            ],
          },
        })
      }

      return new Observable(() => undefined)
    }),
  })

const renderSettings = (isAdmin: boolean) => {
  vi.mocked(authentication.authToken).mockReturnValue('token')

  render(
    <ApolloProvider client={clientFor(isAdmin)}>
      <MemoryRouter>
        <SettingsPage />
      </MemoryRouter>
    </ApolloProvider>
  )
}

test('a user who is not an admin sees the scan queue for their own albums', async () => {
  // The API lets every user see and cancel the jobs for albums they own. The
  // settings page used to show the queue inside the admin-only scanner
  // section, so that was only reachable through a GraphQL client.
  renderSettings(false)

  expect(await screen.findByText('Scan queue')).toBeInTheDocument()
  expect(screen.getByText('Holiday')).toBeInTheDocument()

  // The library-wide controls stay admin-only.
  expect(screen.queryByText('Scan all users')).toBeNull()
})

test('an admin sees the queue once, next to the scanner controls', async () => {
  renderSettings(true)

  expect(await screen.findByText('Scan all users')).toBeInTheDocument()
  expect(await screen.findAllByText('Scan queue')).toHaveLength(1)
})
