import React from 'react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { MockedProvider } from '@apollo/client/testing'

import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react'

import {clearSharePassword, saveSharePassword} from '../../helpers/authentication'

import {
  SHARE_TOKEN_QUERY,
  TokenRoute,
  VALIDATE_TOKEN_PASSWORD_QUERY,
} from './SharePage'

import { SIDEBAR_DOWNLOAD_QUERY } from '../../components/sidebar/SidebarDownloadMedia'
import { SHARE_ALBUM_QUERY } from './AlbumSharePage'
import Message from '../../components/messages/Message'


vi.mock('../../hooks/useScrollPagination')

describe('load correct share page, based on graphql query', () => {
  const token = 'TOKEN123'
  const password = 'PASSWORD-123_@456\\'

  const historyMock = [{ pathname: `/share/${token}` }]

  const graphqlMocks = [
    {
      request: {
        query: VALIDATE_TOKEN_PASSWORD_QUERY,
        variables: {
          token,
          password,
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

  beforeAll(() => {
    saveSharePassword(token, password)
  })

  afterAll(() => {
    clearSharePassword(token)
  })

  test('load media share page', async () => {
    const mediaPageMock = {
      request: {
        query: SHARE_TOKEN_QUERY,
        variables: {
          token,
          password,
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
          <Routes>
            <Route path="/share/:token/*" element={<TokenRoute />} />
          </Routes>
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitForElementToBeRemoved(() => screen.queryByText('Loading...'))

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
            password,
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
            token,
            password,
            limit: 200,
            offset: 0,
            mediaOrderBy: 'date_shot',
            mediaOrderDirection: 'ASC',
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
          <Routes>
            <Route path="/share/:token/*" element={<TokenRoute />} />
          </Routes>
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    await waitForElementToBeRemoved(() => screen.getByText('Loading...'))

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    expect(screen.getByTestId('AlbumSharePage')).toBeInTheDocument()
  })

  test('load subalbum of a shared album', async () => {
    const subalbumID = '456'
    const subalbumHistoryMock = [{ pathname: `/share/${token}/${subalbumID}` }]

    const subalbumPageMocks = [
      {
        request: {
          query: SHARE_TOKEN_QUERY,
          variables: {
            token,
            password,
          },
        },
        result: {
          data: {
            shareToken: {
              token: token,
              album: {
                id: subalbumID,
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
            id: subalbumID,
            token,
            password,
            limit: 200,
            offset: 0,
            mediaOrderBy: 'date_shot',
            mediaOrderDirection: 'ASC',
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
        mocks={[...graphqlMocks, ...subalbumPageMocks]}
        addTypename={false}
        defaultOptions={{
          // disable cache, required to make fragments work
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' },
        }}
      >
        <MemoryRouter initialEntries={subalbumHistoryMock}>
          <Routes>
            <Route path="/share/:token/*" element={<TokenRoute />} />
          </Routes>
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    await waitForElementToBeRemoved(() => screen.getByText('Loading...'))

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    expect(screen.getByTestId('AlbumSharePage')).toBeInTheDocument()
  })

  test(`test the expired`, async () => {
    const expiredValidationMock = {
      request: {
        query: VALIDATE_TOKEN_PASSWORD_QUERY,
        variables: { token: 'TOKEN123', password: "PASSWORD-123_@456\\" },
      },
      error: new Error('share expired'),
    };
    
    render(
      <MockedProvider
        mocks={[expiredValidationMock]}
        addTypename={false}
        defaultOptions={{
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' },
        }}
      >
        <MemoryRouter initialEntries={historyMock}>
          <Routes>
            <Route path="/share/:token/*" element={<TokenRoute />} />
          </Routes>
        </MemoryRouter>
      </MockedProvider>
    )
    expect(screen.getByText('Loading...')).toBeInTheDocument()

    await waitForElementToBeRemoved(() => screen.queryByText('Loading...'))

    expect(screen.getByText(/share expired/i)).toBeInTheDocument()
  })

test(`test the normal`, async () => {

  const shareTokenMock = {
    request: {
      query: SHARE_TOKEN_QUERY,
      variables: { token: 'TOKEN123', password: "PASSWORD-123_@456\\" },
    },
    result: {
      data: {
        shareToken: {
          token: 'TOKEN123',
          album: { id: '1' }, 
          media: null,
        },
      },
    },
  };

  const albumDataMock = {
    request: {
      query: SHARE_ALBUM_QUERY,
      variables: {
        id: '1',
        token: 'TOKEN123',
        password: "PASSWORD-123_@456\\",
        limit: 200,
        offset: 0,
        mediaOrderBy: 'date_shot',
        mediaOrderDirection: 'ASC',
      },
    },
    result: {
      data: {
        album: {
          id: '1',
          title: 'normal_album',
          subAlbums: [],
          thumbnail: { url: '...' },
          media: [],
        },
      },
    },
  };

  render(
    <MockedProvider
      mocks={[...graphqlMocks, shareTokenMock, albumDataMock]}
      addTypename={false}
    >
      <MemoryRouter initialEntries={historyMock}>
        <Routes>
          <Route path="/share/:token/*" element={<TokenRoute />} />
        </Routes>
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Loading...')).toBeInTheDocument()
  await waitForElementToBeRemoved(() => screen.queryByText('Loading...'))

  expect(screen.getByTestId('Layout')).toBeInTheDocument()
  expect(screen.getByTestId('AlbumSharePage')).toBeInTheDocument()
})


})
