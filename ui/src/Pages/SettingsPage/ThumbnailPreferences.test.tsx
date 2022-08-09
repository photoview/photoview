import React from 'react'
import { MockedProvider } from '@apollo/client/testing'

import { render, screen } from '@testing-library/react'

import { ThumbnailFilter } from '../../__generated__/globalTypes'

import ThumbnailPreferences, {
  THUMBNAIL_METHOD_QUERY,
  SET_THUMBNAIL_METHOD_MUTATION,
} from './ThumbnailPreferences'

test('load ThumbnailPreferences', () => {
  const graphqlMocks = [
    {
      request: {
        query: THUMBNAIL_METHOD_QUERY,
      },
      result: {
        data: {
          siteInfo: { method: ThumbnailFilter.NearestNeighbor },
        },
      },
    },
    {
      request: {
        query: SET_THUMBNAIL_METHOD_MUTATION,
        variables: {
          method: ThumbnailFilter.Lanczos,
        },
      },
      result: {
        data: {},
      },
    },
  ]
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
      <ThumbnailPreferences />
    </MockedProvider>
  )

  expect(screen.getByText('Downsampling method')).toBeInTheDocument()
})
