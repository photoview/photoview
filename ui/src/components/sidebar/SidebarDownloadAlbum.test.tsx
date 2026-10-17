import React from 'react'
import { render, screen } from '@testing-library/react'
import SidebarAlbumDownload from './SidebarDownloadAlbum'

vi.mock('../../apolloClient', () => ({ API_ENDPOINT: '/api' }))

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
