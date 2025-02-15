import React from 'react' //React must be in scope when using JSX.
import { screen, waitFor } from '@testing-library/react'
import LoginPage from './LoginPage'
import * as authentication from '../../helpers/authentication'
import { createMemoryHistory } from 'history'
import { mockInitialSetupGraphql } from './loginTestHelpers'
import { renderWithProviders } from '../../helpers/testUtils'
import { act } from '@testing-library/react'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('Login page redirects', () => {
  test('Auth token redirect', async () => {
    authToken.mockImplementation(() => 'some-token')

    const history = createMemoryHistory()
    history.push('/login')

    await act(async () => {
      await renderWithProviders(<LoginPage />, {
        mocks: [],
        history,
      })
    })

    await waitFor(() => {
      expect(history.location.pathname).toBe('/')
    })
  })

  test('Initial setup redirect', async () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory()
    history.push('/login')

    await act(async () => {
      await renderWithProviders(<LoginPage />, {
        mocks: [mockInitialSetupGraphql(true)],
        history,
      })
    })

    await waitFor(() => {
      expect(history.location.pathname).toBe('/initialSetup')
    })
  })
})

describe('Login page', () => {
  test('Render login form', async () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory()
    history.push('/login')

    await act(async () => {
      await renderWithProviders(<LoginPage />, {
        mocks: [mockInitialSetupGraphql(false)],
        history,
      })
    })

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Sign in')).toBeInTheDocument()
  })
})
