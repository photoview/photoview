import React from 'react'
import { MockedProvider } from '@apollo/client/testing'

import { render, screen } from '@testing-library/react'

import {
  THUMBNAIL_METHOD_QUERY,
  SET_THUMBNAIL_METHOD_MUTATION,
  ThumbnailPreferences,
} from './ThumbnailPreferences'

test('load ThumbnailPreferences', () => {
  const graphqlMocks = [
    {
      request: {
        query: THUMBNAIL_METHOD_QUERY,
      },
      result: {
        data: {
          siteInfo: { method: 0 },
        },
      },
    },
    {
      request: {
        query: SET_THUMBNAIL_METHOD_MUTATION,
        variables: {
          method: '5',
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
