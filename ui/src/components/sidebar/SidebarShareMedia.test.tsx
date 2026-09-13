import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SidebarShareMediaButton } from './SidebarDownloadMedia'
import { MediaSidebarMedia } from './MediaSidebar/MediaSidebar'

const media = {
  id: '1',
  title: 'holiday.jpg',
  type: 'Photo',
} as MediaSidebarMedia

const rows = [
  {
    title: 'Original',
    url: 'photo/holiday.jpg',
    width: 1024,
    height: 768,
    fileSize: 1234,
  },
]

// Mimics a download slow enough to outlast the transient user activation: the
// first share attempt is refused, a later one is not.
const refuseUntilSecondAttempt = () => {
  let attempts = 0

  return vi.fn(() => {
    attempts += 1
    if (attempts === 1) {
      const err = new Error('user activation is gone')
      err.name = 'NotAllowedError'

      return Promise.reject(err)
    }

    return Promise.resolve()
  })
}

let downloads = 0
const originalFetch = global.fetch
const originalNavigator = global.navigator

beforeEach(() => {
  downloads = 0

  global.fetch = vi.fn(() => {
    downloads += 1

    return Promise.resolve({
      ok: true,
      status: 200,
      blob: () => Promise.resolve(new Blob(['x'], { type: 'image/jpeg' })),
    })
  }) as unknown as typeof fetch
})

afterEach(() => {
  global.fetch = originalFetch
  Object.defineProperty(global, 'navigator', {
    value: originalNavigator,
    configurable: true,
    writable: true,
  })
})

const renderWithShare = (
  share: ReturnType<typeof vi.fn>,
  canShareFiles: boolean
) => {
  Object.defineProperty(global, 'navigator', {
    value: {
      ...originalNavigator,
      share,
      canShare: () => canShareFiles,
    },
    configurable: true,
    writable: true,
  })

  render(<SidebarShareMediaButton media={media} rows={rows} />)

  return share
}

const setup = (canShareFiles: boolean) =>
  renderWithShare(refuseUntilSecondAttempt(), canShareFiles)

test('a refused file share can be retried without downloading again', async () => {
  const share = setup(true)

  await userEvent.click(screen.getByRole('button'))
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(downloads).toBe(1)

  // The button has to say that a second tap is what is needed - nothing else
  // on screen would tell the user the first one did anything at all.
  const retryButton = await screen.findByRole('button', {
    name: /again/i,
  })

  await userEvent.click(retryButton)
  await waitFor(() => expect(share).toHaveBeenCalledTimes(2))

  // The kept file is the whole point: downloading again would burn the fresh
  // activation exactly as the first attempt did.
  expect(downloads).toBe(1)
})

test('a refused link fallback can be retried too', async () => {
  // A browser that shares links but not files still pays for the download
  // first, so it can lose the activation in the same way.
  const share = setup(false)

  await userEvent.click(screen.getByRole('button'))
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(downloads).toBe(1)

  const retryButton = await screen.findByRole('button', {
    name: /again/i,
  })

  await userEvent.click(retryButton)
  await waitFor(() => expect(share).toHaveBeenCalledTimes(2))
  expect(downloads).toBe(1)
})

test('a failed download still shares the link', async () => {
  global.fetch = vi.fn(() => {
    downloads += 1

    return Promise.reject(new Error('the network went away'))
  }) as unknown as typeof fetch

  const share = renderWithShare(
    vi.fn(() => Promise.resolve()),
    true
  )

  await userEvent.click(screen.getByRole('button'))
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(downloads).toBe(1)

  // No file could be prepared, but the link is still shareable - ending in
  // nothing would be the one outcome worse than sharing less.
  const payload = share.mock.calls[0][0] as Record<string, unknown>
  expect(payload).not.toHaveProperty('files')
  expect(payload.url).toBe(location.href)

  // Nothing was held back, so the button is not asking for a second tap that
  // would only fail the same way.
  expect(screen.getByRole('button')).not.toHaveAccessibleName(/again/i)
})

test('a failed download whose link share is refused retries the link alone', async () => {
  // The worst of both: the download fails, so there is no file to keep, and it
  // was slow enough that the link share standing in for it has no activation
  // left either. Without a memory of that, every further tap would pay for the
  // same failing download before reaching the link again.
  global.fetch = vi.fn(() => {
    downloads += 1

    return Promise.reject(new Error('the network went away'))
  }) as unknown as typeof fetch

  const share = renderWithShare(refuseUntilSecondAttempt(), true)

  await userEvent.click(screen.getByRole('button'))
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(downloads).toBe(1)

  const retryButton = await screen.findByRole('button', { name: /again/i })

  await userEvent.click(retryButton)
  await waitFor(() => expect(share).toHaveBeenCalledTimes(2))

  // The second tap went straight to the link, inside its own activation.
  expect(downloads).toBe(1)
  const payload = share.mock.calls[1][0] as Record<string, unknown>
  expect(payload).not.toHaveProperty('files')
  expect(payload.url).toBe(location.href)

  // And the button drops the retry wording once the share went through.
  await waitFor(() =>
    expect(screen.getByRole('button')).not.toHaveAccessibleName(/again/i)
  )
})
