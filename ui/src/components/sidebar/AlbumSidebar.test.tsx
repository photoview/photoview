import React from 'react'
import { render, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import AlbumSidebar from './AlbumSidebar'
import { SHARE_ALBUM_QUERY } from './Sharing'

import * as authentication from '../../helpers/authentication'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('AlbumSidebar', () => {
  test('render sidebar, unauthorized', () => {
    authToken.mockImplementation(() => null)

    const { getByText, queryByText } = render(
      <AlbumSidebar albumId="30" albumTitle="testingTitle" />
    )

    expect(getByText('testingTitle')).toBeInTheDocument()
    expect(getByText('Download')).toBeInTheDocument()

    expect(queryByText('Sharing options')).not.toBeInTheDocument()
    expect(queryByText('Album cover')).not.toBeInTheDocument()
  })

  test('render sidebar, authorized', async () => {
    authToken.mockImplementation(() => 'token-here')
    const shareMock = {
      request: { query: SHARE_ALBUM_QUERY, variables: { id: '30' } },
      result: {
        data: {
          album: {
            id: '30',
            shares: [
              {
                id: '6',
                token: 'qDSL5I1N',
                hasPassword: false,
                __typename: 'ShareToken',
              },
            ],
            __typename: 'Album',
          },
        },
      },
    }
    const { getByText } = render(
      <MockedProvider mocks={[shareMock]}>
        <AlbumSidebar albumId="30" albumTitle="testingTitle" />
      </MockedProvider>
    )
    await waitFor(() => {
      expect(getByText('testingTitle')).toBeInTheDocument()
      expect(getByText('Download')).toBeInTheDocument()

      expect(getByText('Sharing options')).toBeInTheDocument()

      expect(getByText('Album cover')).toBeInTheDocument()
    })
  })
})
