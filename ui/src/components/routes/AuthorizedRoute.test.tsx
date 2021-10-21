import React from 'react'
import AuthorizedRoute from './AuthorizedRoute'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route } from 'react-router-dom'

import * as authentication from '../../helpers/authentication'

jest.mock('../../helpers/authentication.ts')

const authToken = authentication.authToken as jest.Mock<
  ReturnType<typeof authentication.authToken>
>

describe('AuthorizedRoute component', () => {
  const AuthorizedComponent = () => <div>authorized content</div>

  test('not logged in', async () => {
    authToken.mockImplementation(() => null)

    render(
      <MemoryRouter initialEntries={['/']}>
        <Route path="/login">login redirect</Route>
        <AuthorizedRoute path="/" component={AuthorizedComponent} />
      </MemoryRouter>
    )

    expect(screen.getByText('login redirect')).toBeInTheDocument()
  })

  test('logged in', async () => {
    authToken.mockImplementation(() => 'token-here')

    render(
      <MemoryRouter initialEntries={['/']}>
        <Route path="/login">login redirect</Route>
        <AuthorizedRoute path="/" component={AuthorizedComponent} />
      </MemoryRouter>
    )

    expect(screen.getByText('authorized content')).toBeInTheDocument()
    expect(screen.queryByText('login redirect')).toBeNull()
  })
})
