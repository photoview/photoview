import React from 'react'
import { render, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import AlbumTitle, { ALBUM_PATH_QUERY } from './AlbumTitle'
import { MemoryRouter } from 'react-router'

import * as authentication from '../../helpers/authentication'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

const album = {
  id: '30',
  title: 'testAlbum',
}

describe('AlbumTitle', () => {
  test('render albumTitle - no album given, unauthorized', () => {
    authToken.mockImplementation(() => null)
    const { queryByText } = render(
      <MockedProvider mocks={[]}>
        <AlbumTitle disableLink />
      </MockedProvider>
    )
    expect(queryByText('Album options')).not.toBeInTheDocument()
  })

  test('render albumTitle - album given, disableLink=true, unauthorized', async () => {
    authToken.mockImplementation(() => null)

    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <AlbumTitle album={album} disableLink={true} />
        </MemoryRouter>
      </MockedProvider>
    )
    await waitFor(() => {
      // check if the string "Album options" is rendered / the settings button is visible
      const button = document.querySelector('button')
      const title = button?.getAttribute('title')
      expect(title).toBe('Album options')

      // check if the name of the album is in span
      const span = document.querySelector('h1 > span')
      expect(span?.textContent).toMatch(/testAlbum/)

      // check if the name of the album is not in a link
      const linkSpan = document.querySelector('h1 > link > span')
      expect(linkSpan).toBeNull()
    })
  })

  test('render albumTitle - album given, disableLink=false, unauthorized', async () => {
    authToken.mockImplementation(() => null)

    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <AlbumTitle album={album} disableLink={false} />
        </MemoryRouter>
      </MockedProvider>
    )
    await waitFor(() => {
      // check if the string "Album options" is rendered / the settings button is visible
      const button = document.querySelector('button')
      const title = button?.getAttribute('title')
      expect(title).toBe('Album options')

      // check if the name of the album is in span
      const span = document.querySelector('h1 > span')
      expect(span).toBeNull()

      // check if the name of the album is in a link
      const linkSpan = document.querySelector('h1 > a > span')
      expect(linkSpan?.textContent).toMatch(/testAlbum/)
    })
  })

  test('render albumTitle - album given, disableLink=true, authorized, breadcrumb', async () => {
    authToken.mockImplementation(() => 'token-here')
    const subfolderAlbumMock = {
      request: { query: ALBUM_PATH_QUERY, variables: { id: '30' } },
      result: {
        data: {
          album: {
            id: '30',
            title: 'testAlbum',
            path: [{ id: '2', title: 'demoAlbum', __typename: 'Album' }],
            __typename: 'Album',
          },
        },
      },
    }

    render(
      <MockedProvider mocks={[subfolderAlbumMock]}>
        <MemoryRouter>
          <AlbumTitle album={album} disableLink={true} />
        </MemoryRouter>
      </MockedProvider>
    )
    await waitFor(() => {
      // check if the string "Album options" is rendered / the settings button is visible
      const button = document.querySelector('button')
      const title = button?.getAttribute('title')
      expect(title).toBe('Album options')

      // check if the name of the album is in span
      const span = document.querySelector('h1 > span')
      expect(span?.textContent).toMatch(/testAlbum/)

      // check if the name of the album is not in a link
      const linkSpan = document.querySelector('h1 > link > span')
      expect(linkSpan).toBeNull()

      // check if the breadcrumb is rendered
      const breadcrumb = document.querySelector('nav > ol > li > a')
      expect(breadcrumb?.textContent).toMatch(/demoAlbum/)
    })
  })
})
