import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import {
  SCANNER_QUEUE_STATUS_QUERY,
  CANCEL_ALL_SCAN_JOBS_MUTATION,
  ScannerQueueStatus,
} from './ScannerQueueStatus'
import { ScannerJobStatus } from '../../__generated__/globalTypes'

const queueItem = {
  __typename: 'ScannerQueueItem',
  status: ScannerJobStatus.RUNNING,
  album: {
    __typename: 'Album',
    id: '1',
    title: 'Vacation',
    path: [],
  },
}

test('renders queue items and cancels all jobs', async () => {
  let cancelAllCalled = false

  const graphqlMocks = [
    {
      request: { query: SCANNER_QUEUE_STATUS_QUERY },
      result: { data: { scannerQueueStatus: [queueItem] } },
    },
    {
      request: { query: CANCEL_ALL_SCAN_JOBS_MUTATION },
      result: () => {
        cancelAllCalled = true
        return { data: { cancelAllScanJobs: 1 } }
      },
    },
    {
      request: { query: SCANNER_QUEUE_STATUS_QUERY },
      result: { data: { scannerQueueStatus: [] } },
    },
  ]

  render(
    <MockedProvider
      mocks={graphqlMocks}
      addTypename={false}
      defaultOptions={{
        watchQuery: { fetchPolicy: 'no-cache' },
        query: { fetchPolicy: 'no-cache' },
      }}
    >
      <ScannerQueueStatus />
    </MockedProvider>
  )

  expect(await screen.findByText('Vacation')).toBeInTheDocument()

  await userEvent.click(screen.getByText('Cancel all'))

  await waitFor(() => expect(cancelAllCalled).toBe(true))
})
