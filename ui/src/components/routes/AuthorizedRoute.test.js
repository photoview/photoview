import '@testing-library/jest-dom'

import React from 'react'
import AuthorizedRoute from './AuthorizedRoute'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route } from 'react-router-dom'

import * as authentication from '../../authentication'

describe('AuthorizedRoute component', () => {
  const AuthorizedComponent = () => <div>authorized content</div>

  test('not logged in', async () => {
    authentication.authToken = jest.fn(() => null)

    render(
      <MemoryRouter initialEntries={['/']}>
        <Route path="/login">login redirect</Route>
        <AuthorizedRoute path="/" component={AuthorizedComponent} />
      </MemoryRouter>
    )

    expect(screen.getByText('login redirect')).toBeInTheDocument()
  })

  test('logged in', async () => {
    authentication.authToken = jest.fn(() => 'token-here')

    render(
      <MemoryRouter initialEntries={['/']}>
        <Route path="/login">login redirect</Route>
        <AuthorizedRoute path="/" component={AuthorizedComponent} />
      </MemoryRouter>
    )

    expect(screen.getByText('authorized content')).toBeInTheDocument()
  })
})
