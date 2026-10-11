import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { unstable_HistoryRouter as HistoryRouter } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import * as authentication from '../../helpers/authentication'
import InitialSetupPage from './InitialSetupPage'
import { mockInitialSetupGraphql } from './loginTestHelpers'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('Initial setup page', () => {
  test('Render initial setup form', () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory({
      initialEntries: ['/initialSetup'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(true)]}>
        <HistoryRouter history={history}>
          <InitialSetupPage />
        </HistoryRouter>
      </MockedProvider>
    )

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByLabelText('Photo path')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Setup Photoview')).toBeInTheDocument()
  })

  test('Redirect if auth token is present', async () => {
    authToken.mockImplementation(() => 'some-token')

    const history = createMemoryHistory({
      initialEntries: ['/initialSetup'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(true)]}>
        <HistoryRouter history={history}>
          <InitialSetupPage />
        </HistoryRouter>
      </MockedProvider>
    )

    await waitFor(() => {
      expect(history.location.pathname).toBe('/')
    })
  })

  test('Redirect if not initial setup', async () => {
    authToken.mockImplementation(() => null)

    const history = createMemoryHistory({
      initialEntries: ['/initialSetup'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(false)]}>
        <HistoryRouter history={history}>
          <InitialSetupPage />
        </HistoryRouter>
      </MockedProvider>
    )

    await waitFor(() => {
      expect(history.location.pathname).toBe('/')
    })
  })
  const renderSetupForm = () => {
    authToken.mockImplementation(() => undefined)

    const history = createMemoryHistory({
      initialEntries: ['/initialSetup'],
    })

    render(
      <MockedProvider mocks={[mockInitialSetupGraphql(true)]}>
        <HistoryRouter history={history}>
          <InitialSetupPage />
        </HistoryRouter>
      </MockedProvider>
    )
  }

  test('the admin password is masked and can be revealed', async () => {
    renderSetupForm()

    // It used to be a plain text field, so a typo could not lock the new
    // admin out. The reveal button covers that now without showing the
    // password to anyone looking at the screen.
    const password = screen.getByLabelText('Password')
    expect(password).toHaveAttribute('type', 'password')

    await userEvent.click(screen.getByRole('button', { name: 'Show password' }))
    expect(password).toHaveAttribute('type', 'text')
  })

  test('each empty required field reports its own error', async () => {
    renderSetupForm()

    await userEvent.type(screen.getByLabelText('Username'), 'admin')
    await userEvent.type(screen.getByLabelText('Password'), 'secret')
    await userEvent.click(screen.getByDisplayValue('Setup Photoview'))

    // Only the photo path is missing, and that is the error that has to show
    // - it used to be read from the password field's state instead.
    expect(
      await screen.findByText('Please enter a photo path')
    ).toBeInTheDocument()
    expect(screen.queryByText('Please enter a password')).toBeNull()
  })
})
