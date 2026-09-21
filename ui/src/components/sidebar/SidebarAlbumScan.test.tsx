import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MockedProvider, MockedResponse } from '@apollo/client/testing'
import { SidebarAlbumScan, SCAN_ALBUM_MUTATION } from './SidebarAlbumScan'

const scanMock = (
  albumId: string,
  outcome: Pick<MockedResponse, 'result' | 'error'>
): MockedResponse & { called: () => number } => {
  let calls = 0

  return {
    request: { query: SCAN_ALBUM_MUTATION, variables: { albumId } },
    // Long enough to look at the button while the request is out.
    delay: 100,
    ...(outcome.error
      ? { error: outcome.error }
      : {
          result: () => {
            calls += 1

            return outcome.result as { data: unknown }
          },
        }),
    // Apollo mocks are single-use, so two taps need two entries.
    called: () => calls,
  }
}

const success = {
  result: {
    data: { scanAlbum: { success: true, __typename: 'ScannerResult' } },
  },
}

const button = () =>
  screen.getByRole('button', { name: 'Rescan this album and its sub-albums' })

test('the button scans this album and can be used again once the scan is queued', async () => {
  const first = scanMock('7', success)
  const second = scanMock('7', success)

  render(
    <MockedProvider mocks={[first, second]} addTypename={false}>
      <SidebarAlbumScan id="7" />
    </MockedProvider>
  )

  await userEvent.click(button())
  // Disabled while the request is out, so a second tap cannot queue it twice.
  expect(button()).toBeDisabled()

  await waitFor(() => expect(first.called()).toBe(1))
  // `called` stays true after a mutation unless it is reset - without that
  // the button would never come back.
  await waitFor(() => expect(button()).toBeEnabled())

  await userEvent.click(button())
  await waitFor(() => expect(second.called()).toBe(1))
  await waitFor(() => expect(button()).toBeEnabled())
})

test('a failed scan request gives the button back', async () => {
  render(
    <MockedProvider
      mocks={[scanMock('7', { error: new Error('scanner unavailable') })]}
      addTypename={false}
    >
      <SidebarAlbumScan id="7" />
    </MockedProvider>
  )

  await userEvent.click(button())
  expect(button()).toBeDisabled()

  await waitFor(() => expect(button()).toBeEnabled())
})
