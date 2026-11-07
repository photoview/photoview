import { render, screen, waitFor } from '@testing-library/react'
import React from 'react'
import LoginPage from './LoginPage'
import * as authentication from '../../helpers/authentication'
import { unstable_HistoryRouter as HistoryRouter } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { MockedProvider } from '@apollo/client/testing'
import { mockInitialSetupGraphql } from './loginTestHelpers'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('Login page redirects', () => {
  test('Auth token redirect', async () => {
    authToken.mockImplementation(() => 'some-token')

    const history = createMemoryHistory({
      initialEntries: ['/login'],
    })

    render(
      <MockedProvider mocks={[]}>
        <HistoryRouter history={history}>
          <LoginPage />
        </HistoryRouter>
      </MockedProvider>
    )

    await waitFor(() => {
      expect(history.location.pathname).toBe('/')
    })
  })

  test('Initial setup redirect', async () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory({
      initialEntries: ['/login'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(true)]}>
        <HistoryRouter history={history}>
          <LoginPage />
        </HistoryRouter>
      </MockedProvider>
    )

    await waitFor(() => {
      expect(history.location.pathname).toBe('/initialSetup')
    })
  })
})

describe('Login page', () => {
  test('Render login form', () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory({
      initialEntries: ['/login'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(false)]}>
        <HistoryRouter history={history}>
          <LoginPage />
        </HistoryRouter>
      </MockedProvider>
    )

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Sign in')).toBeInTheDocument()
  })

  test('offers SSO alongside password login when OIDC is enabled', async () => {
    authToken.mockImplementation(() => null)
    const previousFetch = globalThis.fetch
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ enabled: true }),
      })
    )

    try {
      const history = createMemoryHistory({ initialEntries: ['/login'] })
      render(
        <MockedProvider mocks={[mockInitialSetupGraphql(false)]}>
          <HistoryRouter history={history}>
            <LoginPage />
          </HistoryRouter>
        </MockedProvider>
      )

      expect(await screen.findByText('Sign in with SSO')).toBeInTheDocument()
      expect(screen.getByLabelText('Password')).toBeInTheDocument()
    } finally {
      vi.stubGlobal('fetch', previousFetch)
    }
  })
})
