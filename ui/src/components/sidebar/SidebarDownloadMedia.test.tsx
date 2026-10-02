import * as authentication from '../../helpers/authentication'
import { fetchMediaBlob, fetchMediaBlobQuiet } from './SidebarDownloadMedia'

vi.mock('../../helpers/authentication')

const t = ((_key: string, defaultValue: string) =>
  defaultValue) as unknown as Parameters<typeof fetchMediaBlob>[0]

const originalFetch = global.fetch
let requested: string[]

const respondWith = (status: number, body = 'x') => {
  global.fetch = vi.fn((url: string) => {
    requested.push(url)

    return Promise.resolve({
      ok: status >= 200 && status < 300,
      status,
      headers: new Headers(),
      blob: () => Promise.resolve(new Blob([body], { type: 'image/jpeg' })),
    })
  }) as unknown as typeof fetch
}

beforeEach(() => {
  requested = []
  vi.mocked(authentication.authToken).mockReturnValue('token')
  vi.spyOn(console, 'error').mockImplementation(() => undefined)
})

afterEach(() => {
  global.fetch = originalFetch
  window.history.pushState({}, '', '/')
})

describe('a response that is not a success', () => {
  test('never reaches the downloaded file', async () => {
    // fetch resolves for a 404 as happily as for a 200. Without the check,
    // the error page would be saved to disk as if it were the photo.
    respondWith(404, '<html>not found</html>')

    await expect(fetchMediaBlob(t)('photo/holiday.jpg')).resolves.toBeNull()
  })

  test('fails the share instead of sharing an error page', async () => {
    respondWith(403, '<html>forbidden</html>')

    await expect(fetchMediaBlobQuiet('photo/holiday.jpg')).rejects.toThrow(
      '403'
    )
  })

  test('still lets a successful response through', async () => {
    respondWith(200)

    await expect(
      fetchMediaBlobQuiet('photo/holiday.jpg')
    ).resolves.toBeInstanceOf(Blob)
  })
})

describe('media behind a share link', () => {
  test('carries the share token when nobody is logged in', async () => {
    // An anonymous visitor of /share/:token has no session; the media route
    // authorizes them through the token instead.
    vi.mocked(authentication.authToken).mockReturnValue(undefined)
    window.history.pushState({}, '', '/share/abc123/photo')
    respondWith(200)

    await fetchMediaBlobQuiet('photo/holiday.jpg')

    expect(new URL(requested[0]).searchParams.get('token')).toBe('abc123')
  })

  test('leaves the token off for a logged-in user', async () => {
    window.history.pushState({}, '', '/share/abc123/photo')
    respondWith(200)

    await fetchMediaBlobQuiet('photo/holiday.jpg')

    expect(new URL(requested[0]).searchParams.has('token')).toBe(false)
  })
})
