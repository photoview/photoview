import Cookies from 'js-cookie'

import {
  saveTokenCookie,
  clearTokenCookie,
  authToken,
  saveSharePassword,
  clearSharePassword,
  getSharePassword,
} from './authentication'

function resetCookies() {
  Object.keys(Cookies.get()).forEach(function (cookieName) {
    Cookies.remove(cookieName)
  })
}

describe('helpers/authentication', () => {
  afterEach(() => {
    resetCookies()
  })

  test('auth token cookie operations', function () {
    const AUTH_TOKEN = 'MOCK_TOKEN_1234'

    saveTokenCookie(AUTH_TOKEN)

    expect(document.cookie).toEqual(`auth-token=${AUTH_TOKEN}`)

    expect(authToken()).toEqual(AUTH_TOKEN)

    clearTokenCookie()

    expect(authToken()).toBeUndefined()
    expect(document.cookie).toEqual('')
  })

  test('share token cookie operations', function () {
    const SHARES = {
      MOCK_SHARE_TOKEN1: 'simple',
      MOCK_SHARE_TOKEN2: 'has-some-special-characters-!23@123',
      MOCK_SHARE_TOKEN3: '\\',
    }

    for (const shareToken in SHARES) {
      const password = SHARES[shareToken as keyof typeof SHARES]
      saveSharePassword(shareToken, password)
    }

    expect(document.cookie).toEqual(
      'share-token-pw-MOCK_SHARE_TOKEN1=simple; share-token-pw-MOCK_SHARE_TOKEN2=has-some-special-characters-!23@123; share-token-pw-MOCK_SHARE_TOKEN3=%5C'
    )

    expect(getSharePassword('MOCK_SHARE_TOKEN1')).toEqual(
      SHARES.MOCK_SHARE_TOKEN1
    )
    expect(getSharePassword('MOCK_SHARE_TOKEN2')).toEqual(
      SHARES.MOCK_SHARE_TOKEN2
    )
    expect(getSharePassword('MOCK_SHARE_TOKEN3')).toEqual(
      SHARES.MOCK_SHARE_TOKEN3
    )

    clearSharePassword('MOCK_SHARE_TOKEN1')
    clearSharePassword('MOCK_SHARE_TOKEN2')
    clearSharePassword('MOCK_SHARE_TOKEN3')

    expect(getSharePassword('MOCK_SHARE_TOKEN1')).toBeUndefined()
    expect(getSharePassword('MOCK_SHARE_TOKEN2')).toBeUndefined()
    expect(getSharePassword('MOCK_SHARE_TOKEN3')).toBeUndefined()
    expect(document.cookie).toEqual('')
  })
})
