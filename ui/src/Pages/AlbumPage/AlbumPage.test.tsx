import {
  ApolloClient,
  ApolloLink,
  ApolloProvider,
  InMemoryCache,
  Observable,
} from '@apollo/client'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import AlbumPage from './AlbumPage'
import * as authentication from '../../helpers/authentication'

vi.mock('../../hooks/useScrollPagination')
vi.mock('../../helpers/authentication')

test('AlbumPage renders', () => {
  render(
    <MockedProvider mocks={[]}>
      <MemoryRouter initialEntries={['/album/1']}>
        <Routes>
          <Route path="/album/:id" element={<AlbumPage />} />
        </Routes>
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Sort')).toBeInTheDocument()
  expect(screen.getByLabelText('Sort direction')).toBeInTheDocument()

  screen.debug()
})

describe('the mobile album tree entry point', () => {
  // Answers only the tree preference. Every other query - the album itself -
  // stays pending, which keeps the page rendering its header without having
  // to reproduce a whole album response.
  const clientWith = (showAlbumTree: boolean) =>
    new ApolloClient({
      cache: new InMemoryCache(),
      link: new ApolloLink(operation =>
        operation.operationName === 'layoutAlbumTreePreferenceQuery'
          ? Observable.of({
              data: {
                myUserPreferences: {
                  __typename: 'UserPreferences',
                  id: '1',
                  showAlbumTree,
                },
              },
            })
          : new Observable(() => undefined)
      ),
    })

  const renderAlbum = (showAlbumTree: boolean) => {
    vi.mocked(authentication.authToken).mockReturnValue('token')

    render(
      <ApolloProvider client={clientWith(showAlbumTree)}>
        <MemoryRouter initialEntries={['/album/1']}>
          <Routes>
            <Route path="/album/:id" element={<AlbumPage />} />
          </Routes>
        </MemoryRouter>
      </ApolloProvider>
    )
  }

  test('is there inside an album, not only on the album list', async () => {
    // The tree exists to move between folders, so on a phone losing its
    // button the moment one opens an album defeats the purpose.
    renderAlbum(true)

    expect(
      await screen.findByRole('button', { name: 'Browse album tree' })
    ).toBeInTheDocument()
  })

  test('stays away while the tree is switched off', async () => {
    renderAlbum(false)

    expect(await screen.findByText('Sort')).toBeInTheDocument()
    await new Promise(resolve => setTimeout(resolve, 50))
    expect(
      screen.queryByRole('button', { name: 'Browse album tree' })
    ).toBeNull()
  })
})
