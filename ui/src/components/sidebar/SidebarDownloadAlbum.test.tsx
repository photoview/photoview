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
