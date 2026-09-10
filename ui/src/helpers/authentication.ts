import Cookies from 'js-cookie'

const COOKIE_DEFAULT_OPTIONS = {
  path: '/',
  sameSite: 'Lax',
} as Partial<Cookies.CookieAttributes>

const AUTH_TOKEN_MAX_AGE_IN_DAYS = 14
const AUTH_TOKEN_COOKIE_NAME = 'auth-token'

const SHARE_TOKEN_COOKIE_NAME = (shareToken: string) =>
  `share-token-pw-${shareToken}`

// Session storage survives a header remount (each page wraps itself in its
// own <Layout>) but must not survive an account switch on the same
// browser tab, or the next account could inherit the previous one's typed
// search query.
const SEARCH_QUERY_STORAGE_KEY = 'searchbar.query'

export function readStoredSearchQuery(): string {
  try {
    return sessionStorage.getItem(SEARCH_QUERY_STORAGE_KEY) ?? ''
  } catch {
    return ''
  }
}

export function writeStoredSearchQuery(query: string) {
  try {
    sessionStorage.setItem(SEARCH_QUERY_STORAGE_KEY, query)
  } catch {
    // Ignore storage errors (e.g. private browsing with storage disabled)
  }
}

function clearStoredSearchQuery() {
  try {
    sessionStorage.removeItem(SEARCH_QUERY_STORAGE_KEY)
  } catch {
    // Ignore storage errors (e.g. private browsing with storage disabled)
  }
}

export function saveTokenCookie(token: string) {
  const options = {
    ...COOKIE_DEFAULT_OPTIONS,
    expires: AUTH_TOKEN_MAX_AGE_IN_DAYS,
  }

  Cookies.set(AUTH_TOKEN_COOKIE_NAME, token, options)
}

export function clearTokenCookie() {
  Cookies.remove(AUTH_TOKEN_COOKIE_NAME)
  clearStoredSearchQuery()
}

export function authToken() {
  return Cookies.get(AUTH_TOKEN_COOKIE_NAME)
}

export function saveSharePassword(shareToken: string, password: string) {
  const cookieName = SHARE_TOKEN_COOKIE_NAME(shareToken)

  Cookies.set(cookieName, password, COOKIE_DEFAULT_OPTIONS)
}

export function clearSharePassword(shareToken: string) {
  const cookieName = SHARE_TOKEN_COOKIE_NAME(shareToken)

  Cookies.remove(cookieName)
}

export function getSharePassword(shareToken: string) {
  const cookieName = SHARE_TOKEN_COOKIE_NAME(shareToken)

  return Cookies.get(cookieName)
}
