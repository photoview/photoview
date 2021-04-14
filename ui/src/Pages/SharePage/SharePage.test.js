import '@testing-library/jest-dom'

import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import { MockedProvider } from '@apollo/client/testing'

import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react'

import SharePage, {
  SHARE_TOKEN_QUERY,
  VALIDATE_TOKEN_PASSWORD_QUERY,
} from './SharePage'

import { SIDEBAR_DOWNLOAD_QUERY } from '../../components/sidebar/SidebarDownload'
import { SHARE_ALBUM_QUERY } from './AlbumSharePage'

require('../../localization').setupLocalization()

describe('load correct share page, based on graphql query', () => {
  const token = 'TOKEN123'

  const matchMock = {
    url: `/share`,
    params: {},
    path: `/share`,
    isExact: false,
  }

  const historyMock = [{ pathname: `/share/${token}` }]

  const graphqlMocks = [
    {
      request: {
        query: VALIDATE_TOKEN_PASSWORD_QUERY,
        variables: {
          token,
          password: null,
        },
      },
      result: {
        data: {
          shareTokenValidatePassword: true,
        },
      },
    },
    {
      request: {
        query: SIDEBAR_DOWNLOAD_QUERY,
        variables: {
          mediaId: '1',
        },
      },
      result: {
        data: {
          media: {
            id: '1',
            downloads: [],
          },
        },
      },
    },
  ]

  test('load media share page', async () => {
    const mediaPageMock = {
      request: {
        query: SHARE_TOKEN_QUERY,
        variables: {
          token,
          password: null,
        },
      },
      result: {
        data: {
          shareToken: {
            token: token,
            album: null,
            media: {
              id: '1',
              title: 'shared_image.jpg',
              type: 'Photo',
              highRes: {
                url: 'https://example.com/shared_image.jpg',
              },
            },
          },
        },
      },
    }

    render(
      <MockedProvider
        mocks={[...graphqlMocks, mediaPageMock]}
        addTypename={false}
        defaultOptions={{
          // disable cache, required to make fragments work
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' },
        }}
      >
        <MemoryRouter initialEntries={historyMock}>
          <SharePage match={matchMock} />
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitForElementToBeRemoved(() => screen.getByText('Loading...'))

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    expect(screen.getByTestId('MediaSharePage')).toBeInTheDocument()
  })

  test('load album share page', async () => {
    const albumPageMock = [
      {
        request: {
          query: SHARE_TOKEN_QUERY,
          variables: {
            token,
            password: null,
          },
        },
        result: {
          data: {
            shareToken: {
              token: token,
              album: {
                id: '1',
              },
              media: null,
            },
          },
        },
      },
      {
        request: {
          query: SHARE_ALBUM_QUERY,
          variables: {
            id: '1',
            token: token,
            password: null,
            limit: 200,
            offset: 0,
          },
        },
        result: {
          data: {
            album: {
              id: '1',
              title: 'album_title',
              subAlbums: [],
              thumbnail: {
                url: 'https://photoview.example.com/album_thumbnail.jpg',
              },
              media: [],
            },
          },
        },
      },
    ]

    render(
      <MockedProvider
        mocks={[...graphqlMocks, ...albumPageMock]}
        addTypename={false}
        defaultOptions={{
          // disable cache, required to make fragments work
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' },
        }}
      >
        <MemoryRouter initialEntries={historyMock}>
          <SharePage match={matchMock} />
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitForElementToBeRemoved(() => screen.getByText('Loading...'))

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    expect(screen.getByTestId('AlbumSharePage')).toBeInTheDocument()
  })
})
