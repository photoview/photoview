import { afterEach, describe, expect, test, vi } from 'vitest'
import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import SidebarAlbumDownload from './SidebarDownloadAlbum'

const API_ENDPOINT = 'http://localhost:4001/api'
const ALBUM_ID = '12'

let mockApiEndpoint = API_ENDPOINT

vi.mock('../../apolloClient', () => ({
  get API_ENDPOINT() {
    return mockApiEndpoint
  },
}))

describe('SidebarAlbumDownload', () => {
  const originalLocation = globalThis.location

  afterEach(() => {
    vi.stubGlobal('location', originalLocation)
  })

  test.each([
    ['without a trailing slash', API_ENDPOINT],
    ['with a trailing slash', `${API_ENDPOINT}/`],
  ])(
    'joins the album download URL for an API endpoint %s',
    (_, apiEndpoint) => {
      mockApiEndpoint = apiEndpoint
      const location = { href: '' }
      vi.stubGlobal('location', location)

      render(<SidebarAlbumDownload albumID={ALBUM_ID} />)
      fireEvent.click(screen.getByText('Originals'))

      expect(location.href).toBe(
        `${API_ENDPOINT}/download/album/${ALBUM_ID}/original`
      )
    }
  )
})

test('each album download is a button named after what it downloads', () => {
  render(<SidebarAlbumDownload albumID="7" />)

  for (const name of [
    'Download Thumbnails',
    'Download High resolutions',
    'Download Originals',
    'Download Converted videos',
  ]) {
    expect(screen.getByRole('button', { name })).toBeInTheDocument()
  }

  // The rows only hold the buttons; a row that is a button of its own would
  // be announced as a table row and nothing else.
  for (const row of screen.getAllByRole('row')) {
    expect(row).not.toHaveAttribute('tabindex')
  }

  // Whole-album zips are not sent to other apps: they would have to be
  // downloaded into memory first.
  expect(screen.queryByRole('button', { name: /send/i })).toBeNull()
})
