import React from 'react'
import { MockedProvider } from '@apollo/client/testing'

import { render, screen } from '@testing-library/react'

import {
  CONCURRENT_WORKERS_QUERY,
  SET_CONCURRENT_WORKERS_MUTATION,
  ScannerConcurrentWorkers,
} from './ScannerConcurrentWorkers'

test('load ScannerConcurrentWorkers', () => {
  const graphqlMocks = [
    {
      request: {
        query: CONCURRENT_WORKERS_QUERY,
      },
      result: {
        data: {
          siteInfo: { concurrentWorkers: 3 },
        },
      },
    },
    {
      request: {
        query: SET_CONCURRENT_WORKERS_MUTATION,
        variables: {
          workers: '1',
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
      <ScannerConcurrentWorkers />
    </MockedProvider>
  )

  expect(screen.getByText('Scanner concurrent workers')).toBeInTheDocument()
})
