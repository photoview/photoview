import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SidebarDownloadTable } from './SidebarDownloadMedia'
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

  render(<SidebarDownloadTable media={media} rows={rows} />)

  return share
}

const sendButton = () => screen.getByRole('button', { name: /send/i })

const setup = (canShareFiles: boolean) =>
  renderWithShare(refuseUntilSecondAttempt(), canShareFiles)

test('a refused file share can be retried without downloading again', async () => {
  const share = setup(true)

  await userEvent.click(sendButton())
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

  await userEvent.click(sendButton())
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

  await userEvent.click(sendButton())
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(downloads).toBe(1)

  // No file could be prepared, but the link is still shareable - ending in
  // nothing would be the one outcome worse than sharing less.
  const payload = share.mock.calls[0][0] as Record<string, unknown>
  expect(payload).not.toHaveProperty('files')
  expect(payload.url).toBe(location.href)

  // Nothing was held back, so the button is not asking for a second tap that
  // would only fail the same way.
  expect(sendButton()).not.toHaveAccessibleName(/again/i)
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

  await userEvent.click(sendButton())
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
  await waitFor(() => expect(sendButton()).not.toHaveAccessibleName(/again/i))
})

test('a browser without canShare gets one share attempt per tap', async () => {
  // No canShare at all: the link is the only thing this browser can take, so
  // there is no file fallback left to try. Reaching for one anyway would open
  // the sheet twice for a single tap.
  const share = refuseUntilSecondAttempt()
  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share },
    configurable: true,
    writable: true,
  })

  render(<SidebarDownloadTable media={media} rows={rows} />)

  await userEvent.click(sendButton())
  const retryButton = await screen.findByRole('button', { name: /again/i })
  expect(share).toHaveBeenCalledTimes(1)
  expect(downloads).toBe(0)

  await userEvent.click(retryButton)
  await waitFor(() => expect(share).toHaveBeenCalledTimes(2))
  expect(downloads).toBe(0)
})

test('cancelling the share sheet ends the share without a retry or a fallback', async () => {
  // The user closing the sheet is a decision, not a failure: nothing should be
  // retried, and the link must not be offered in the file's place.
  const aborted = Object.assign(new Error('share cancelled'), {
    name: 'AbortError',
  })
  const share = renderWithShare(
    vi.fn(() => Promise.reject(aborted)),
    true
  )

  await userEvent.click(sendButton())
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))

  // Give a fallback a chance to fire if there were one.
  await new Promise(resolve => setTimeout(resolve, 50))
  expect(share).toHaveBeenCalledTimes(1)
  expect(sendButton()).not.toHaveAccessibleName(/again/i)
})

test('there is no send column where the browser cannot share', () => {
  // A navigator without share(): spreading the real one would carry jsdom's
  // share over, so the stand-in is built from the one property the component
  // might otherwise read.
  Object.defineProperty(global, 'navigator', {
    value: { userAgent: originalNavigator.userAgent },
    configurable: true,
    writable: true,
  })

  render(<SidebarDownloadTable media={media} rows={rows} />)

  expect(screen.queryByRole('button', { name: /send/i })).toBeNull()
  // The column goes with the button, header included, so the table keeps as
  // many columns in its header as in its rows.
  expect(screen.getAllByRole('columnheader')).toHaveLength(4)
  expect(screen.getAllByRole('cell')).toHaveLength(4)
  expect(
    screen.getByRole('button', { name: 'Download Original' })
  ).toBeInTheDocument()
})

test('the send column has a cell in every row and a header of its own', () => {
  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share: vi.fn(), canShare: () => true },
    configurable: true,
    writable: true,
  })

  render(
    <SidebarDownloadTable
      media={media}
      rows={[rows[0], { ...rows[0], title: 'Small', url: 'photo/small.jpg' }]}
    />
  )

  expect(screen.getAllByRole('columnheader')).toHaveLength(5)
  expect(screen.getAllByRole('cell')).toHaveLength(10)
  expect(screen.getAllByRole('button', { name: /send/i })).toHaveLength(2)
})

test('there is nothing to send for a media without downloads', () => {
  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share: vi.fn(), canShare: () => true },
    configurable: true,
    writable: true,
  })

  render(<SidebarDownloadTable media={media} rows={[]} />)

  expect(screen.queryByRole('button')).toBeNull()
})

test('each row sends its own variant, not a picked one', async () => {
  // Photoview offers every variant for download on purpose; which one to send
  // is the user's choice, so the tapped row is the one that gets shared.
  const share = vi.fn(() => Promise.resolve())
  const requested: string[] = []
  global.fetch = vi.fn((url: string) => {
    requested.push(url)

    return Promise.resolve({
      ok: true,
      status: 200,
      blob: () => Promise.resolve(new Blob(['x'], { type: 'image/jpeg' })),
    })
  }) as unknown as typeof fetch

  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share, canShare: () => true },
    configurable: true,
    writable: true,
  })

  render(
    <SidebarDownloadTable
      media={media}
      rows={[rows[0], { ...rows[0], title: 'Small', url: 'photo/small.jpg' }]}
    />
  )

  await userEvent.click(
    screen.getByRole('button', { name: 'Send Small to another app' })
  )
  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))

  expect(requested).toHaveLength(1)
  expect(requested[0]).toMatch(/photo\/small\.jpg$/)
  const [[shared]] = share.mock.calls as unknown as [[{ files: File[] }]]
  expect(shared.files[0].name).toBe('small.jpg')
})

test('a second tap is asked for below the table, naming the variant', async () => {
  setup(true)

  expect(screen.getByRole('status')).toBeEmptyDOMElement()

  await userEvent.click(sendButton())

  await waitFor(() =>
    expect(screen.getByRole('status')).toHaveTextContent(
      'Tap again to send Original to another app'
    )
  )

  await userEvent.click(sendButton())
  await waitFor(() => expect(screen.getByRole('status')).toBeEmptyDOMElement())
})

test('the name downloads its variant, and the row itself is not a button', async () => {
  const requested: string[] = []
  global.fetch = vi.fn((url: string) => {
    requested.push(url)

    return Promise.resolve({
      ok: true,
      status: 200,
      headers: new Headers(),
      blob: () => Promise.resolve(new Blob(['x'], { type: 'image/jpeg' })),
    })
  }) as unknown as typeof fetch
  const createObjectURL = vi.fn(() => 'blob:photo')
  window.URL.createObjectURL = createObjectURL
  window.URL.revokeObjectURL = vi.fn()

  render(
    <SidebarDownloadTable
      media={media}
      rows={[rows[0], { ...rows[0], title: 'Small', url: 'photo/small.jpg' }]}
    />
  )

  for (const row of screen.getAllByRole('row')) {
    expect(row).not.toHaveAttribute('tabindex')
  }

  await userEvent.click(screen.getByRole('button', { name: 'Download Small' }))
  await waitFor(() => expect(createObjectURL).toHaveBeenCalledTimes(1))

  expect(requested).toHaveLength(1)
  expect(requested[0]).toMatch(/photo\/small\.jpg$/)
})

test('the shared file is named after the url path, without a query string', async () => {
  // A media url can arrive with a share token attached. The filename the share
  // sheet shows must not carry it, and neither must the file's extension.
  const share = vi.fn(() => Promise.resolve())

  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share, canShare: () => true },
    configurable: true,
    writable: true,
  })

  render(
    <SidebarDownloadTable
      media={media}
      rows={[{ ...rows[0], url: 'photo/holiday.jpg?token=abc#preview' }]}
    />
  )

  await userEvent.click(sendButton())

  await waitFor(() => expect(share).toHaveBeenCalled())

  const [[shared]] = share.mock.calls as unknown as [[{ files: File[] }]]
  expect(shared.files[0].name).toBe('holiday.jpg')
})

test('a browser that shares links but not files gets one share attempt per tap', async () => {
  // canShare() says no to the file, so the link is shared instead - and when
  // that fails for a reason other than a lost activation, there is nothing
  // left to try. Sharing it twice would spend an activation that is gone.
  const share = vi.fn(() => Promise.reject(new Error('the sheet went away')))

  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share, canShare: () => false },
    configurable: true,
    writable: true,
  })

  render(<SidebarDownloadTable media={media} rows={rows} />)

  await userEvent.click(sendButton())

  await waitFor(() => expect(share).toHaveBeenCalledTimes(1))
  expect(share).toHaveBeenCalledTimes(1)
})

test('a link shared as the fallback leaves no retry state behind', async () => {
  // The file share fails for a reason that is not a lost activation, so the
  // link stands in for it and succeeds. That is a finished share: the next tap
  // has to start from scratch rather than reuse the file kept for a retry.
  let attempts = 0
  const share = vi.fn(() => {
    attempts += 1
    // The file share is the first call of each tap, the link share the second.
    return attempts === 1
      ? Promise.reject(new Error('the sheet refused the file'))
      : Promise.resolve()
  })

  Object.defineProperty(global, 'navigator', {
    value: { ...originalNavigator, share, canShare: () => true },
    configurable: true,
    writable: true,
  })

  render(<SidebarDownloadTable media={media} rows={rows} />)

  await userEvent.click(sendButton())
  await waitFor(() => expect(share).toHaveBeenCalledTimes(2))

  // No retry asked for, and the second tap prepares the file again.
  expect(sendButton()).not.toHaveAccessibleName(/again/i)
  expect(downloads).toBe(1)

  await userEvent.click(sendButton())
  await waitFor(() => expect(downloads).toBe(2))
})
