import { vi, describe, test, expect } from 'vitest'
import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { SidebarAlbumShare, SET_EXPIRE_MUTATION, SHARE_ALBUM_QUERY } from './Sharing'

// =========================================================
// Key modification: Mock this layer directly to bypass js-cookie.
// This prevents the test runner from reading the authentication.ts file,
// avoiding the "Failed to load js-cookie" error.
// =========================================================
vi.mock('../../helpers/authentication', () => ({
  authToken: () => 'mock_token',
  getSharePassword: () => 'mock_pass',
  saveSharePassword: vi.fn(),
  clearSharePassword: vi.fn(),
}))

// Mock the translation module to prevent interference
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback: string) => fallback,
    i18n: { language: 'en' },
  }),
}))

const MOCK_ALBUM_ID = 'album-123'
const MOCK_TOKEN = 'token-abc'

const mockShareNoExpire = {
  id: 'share-1',
  token: MOCK_TOKEN,
  hasPassword: false,
  expire: null,
  __typename: 'ShareToken'
}

const mockShareWithExpire = {
  ...mockShareNoExpire,
  expire: '2025-01-01T00:00:00Z'
}

const getSharesQueryMock = {
  request: {
    query: SHARE_ALBUM_QUERY,
    variables: { id: MOCK_ALBUM_ID },
  },
  result: {
    data: {
      album: {
        id: MOCK_ALBUM_ID,
        __typename: 'Album',
        shares: [mockShareNoExpire],
      },
    },
  },
}

const setExpireMutationMock = {
  request: {
    query: SET_EXPIRE_MUTATION,
    variables: {
      token: MOCK_TOKEN,
      expire: expect.stringContaining('2026-10-01'),
    },
  },
  result: {
    data: {
      setExpireShareToken: {
        token: MOCK_TOKEN,
        __typename: 'ShareToken',
      },
    },
  },
}

const clearExpireMutationMock = {
  request: {
    query: SET_EXPIRE_MUTATION,
    variables: {
      token: MOCK_TOKEN,
      expire: null,
    },
  },
  result: {
    data: {
      setExpireShareToken: {
        token: MOCK_TOKEN,
        __typename: 'ShareToken',
      },
    },
  },
}

describe('Sidebar Sharing Expiration', () => {
  test('User can open popover, enable expiration, set date, and save', async () => {
    render(
      <MockedProvider mocks={[getSharesQueryMock, setExpireMutationMock]} addTypename={false}>
        <SidebarAlbumShare id={MOCK_ALBUM_ID} />
      </MockedProvider>
    )

    await waitFor(() => expect(screen.getByText(/Public Link/i)).toBeInTheDocument())

    const moreBtn = screen.getByTitle('More')
    fireEvent.click(moreBtn)

    const checkbox = await screen.findByLabelText(/Expiration date/i)
    fireEvent.click(checkbox)

    const inputs = screen.getAllByRole('textbox')

    // Filter for the enabled input field (the password field is disabled by default)
    const dateInput = inputs.find(input => !input.hasAttribute('disabled'))

    if (!dateInput) {
      throw new Error('Could not find the enabled date input field')
    }

    // Interact with the input
    fireEvent.change(dateInput, { target: { value: '2026-10-01' } })
    fireEvent.keyDown(dateInput, { key: 'Enter', code: 'Enter', charCode: 13 })
  })

  test('User can clear the expiration date by unchecking the box', async () => {
    const mockWithExpireData = JSON.parse(JSON.stringify(getSharesQueryMock))
    mockWithExpireData.result.data.album.shares = [mockShareWithExpire]

    render(
      <MockedProvider mocks={[mockWithExpireData, clearExpireMutationMock]} addTypename={false}>
        <SidebarAlbumShare id={MOCK_ALBUM_ID} />
      </MockedProvider>
    )

    await waitFor(() => expect(screen.getByText(/Public Link/i)).toBeInTheDocument())

    fireEvent.click(screen.getByTitle('More'))

    const checkbox = await screen.findByLabelText(/Expiration date/i)
    expect(checkbox).toBeChecked()

    fireEvent.click(checkbox)

    await waitFor(() => expect(checkbox).not.toBeChecked())
  })
})