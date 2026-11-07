import { afterEach, describe, expect, test, vi } from 'vitest'
import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { MediaType } from '../../__generated__/globalTypes'
import { MediaSidebarMedia } from './MediaSidebar/MediaSidebar'
import SidebarMediaDownload from './SidebarDownloadMedia'

vi.mock('../../helpers/authentication', () => ({
  authToken: () => 'mock_token',
}))

const API_HOST = 'http://localhost:4001'
const MEDIA_PATH = '/api/photo/photo_abc.jpg'

const mediaWithDownload = (url: string): MediaSidebarMedia => ({
  __typename: 'Media',
  id: '1',
  type: MediaType.Photo,
  downloads: [
    {
      __typename: 'MediaDownload',
      title: 'Original',
      mediaUrl: {
        __typename: 'MediaURL',
        url,
        width: 800,
        height: 600,
        fileSize: 1024,
      },
    },
  ],
})

describe('SidebarMediaDownload', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    vi.stubGlobal('fetch', originalFetch)
    vi.restoreAllMocks()
  })

  test.each([
    [
      'an absolute API URL',
      `${API_HOST}${MEDIA_PATH}`,
      `${API_HOST}${MEDIA_PATH}`,
    ],
    ['a relative API path', MEDIA_PATH, `${location.origin}${MEDIA_PATH}`],
  ])('fetches the media from %s', async (_, url, expectedUrl) => {
    const fetch = vi.fn().mockResolvedValue({
      headers: new Headers(),
      blob: () => Promise.resolve(new Blob(['image'])),
    })
    vi.stubGlobal('fetch', fetch)
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:photo')
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined)
    const click = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(() => undefined)

    render(
      <MockedProvider mocks={[]}>
        <SidebarMediaDownload media={mediaWithDownload(url)} />
      </MockedProvider>
    )
    fireEvent.click(screen.getByText('Original'))

    expect(fetch).toHaveBeenCalledWith(expectedUrl, { credentials: 'include' })
    await waitFor(() => expect(click).toHaveBeenCalled())
  })
})
