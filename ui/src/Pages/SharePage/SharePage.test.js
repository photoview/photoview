import '@testing-library/jest-dom'

import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import { MockedProvider } from '@apollo/client/testing'
// import { create } from 'react-test-renderer'
import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react'

import SharePage, {
  SHARE_TOKEN_QUERY,
  VALIDATE_TOKEN_PASSWORD_QUERY,
} from './SharePage'

import { MAPBOX_QUERY } from '../../Layout'
import { SIDEBAR_DOWNLOAD_QUERY } from '../../components/sidebar/SidebarDownload'

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
              id: 1,
              title: 'shared_image.jpg',
              type: 'photo',
              highRes: {
                url: 'https://example.com/shared_image.jpg',
              },
            },
          },
        },
      },
    },
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
        query: MAPBOX_QUERY,
      },
      result: {
        data: {
          mapboxToken: null,
        },
      },
    },
    {
      request: {
        query: SIDEBAR_DOWNLOAD_QUERY,
        variables: {
          mediaId: 1,
        },
      },
      result: {
        data: {
          media: {
            id: 1,
            downloads: [],
          },
        },
      },
    },
  ]

  test('load media page', async () => {
    render(
      <MockedProvider
        mocks={graphqlMocks}
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

    expect(screen.getByTestId('MediaSharePage')).toBeInTheDocument()
  })
})
