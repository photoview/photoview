import React from 'react'
import AuthorizedRoute, { useIsAdmin } from './AuthorizedRoute'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'

import * as authentication from '../../helpers/authentication'
import { MockedProvider } from '@apollo/client/testing'
import { ADMIN_QUERY } from '../layout/Layout'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('AuthorizedRoute component', () => {
  const AuthorizedComponent = () => <div>authorized content</div>

  test('not logged in', () => {
    authToken.mockImplementation(() => null)

    render(
      <MemoryRouter initialEntries={['/authorized']}>
        <Routes>
          <Route index element={<div>redirect</div>} />
          <Route
            path="/authorized"
            element={
              <AuthorizedRoute>
                <AuthorizedComponent />
              </AuthorizedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('redirect')).toBeInTheDocument()
    expect(screen.queryByText('authorized content')).toBeNull()
  })

  test('logged in', () => {
    authToken.mockImplementation(() => 'token-here')

    render(
      <MemoryRouter initialEntries={['/authorized']}>
        <Routes>
          <Route index element={<div>redirect</div>} />
          <Route
            path="/authorized"
            element={
              <AuthorizedRoute>
                <AuthorizedComponent />
              </AuthorizedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('authorized content')).toBeInTheDocument()
    expect(screen.queryByText('redirect')).toBeNull()
  })
})

describe('useIsAdmin hook', () => {
  const graphqlMock = (admin: boolean) => ({
    request: {
      query: ADMIN_QUERY,
    },
    result: {
      data: {
        myUser: {
          admin,
        },
      },
    },
  })

  test('not logged in', async () => {
    authToken.mockImplementation(() => null)

    const TestComponent = () => {
      const isAdmin = useIsAdmin()
      return isAdmin ? <div>is admin</div> : <div>not admin</div>
    }

    render(
      <MockedProvider mocks={[graphqlMock(true)]}>
        <TestComponent />
      </MockedProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('not admin')).toBeInTheDocument()
    })
  })

  test('not admin', async () => {
    authToken.mockImplementation(() => 'token-here')

    const TestComponent = () => {
      const isAdmin = useIsAdmin()
      return isAdmin ? <div>is admin</div> : <div>not admin</div>
    }

    render(
      <MockedProvider mocks={[graphqlMock(false)]}>
        <TestComponent />
      </MockedProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('not admin')).toBeInTheDocument()
    })
  })

  test('is admin', async () => {
    authToken.mockImplementation(() => 'token-here')

    const TestComponent = () => {
      const isAdmin = useIsAdmin()
      return isAdmin ? <div>is admin</div> : <div>not admin</div>
    }

    render(
      <MockedProvider mocks={[graphqlMock(true)]}>
        <TestComponent />
      </MockedProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('is admin')).toBeInTheDocument()
    })
  })
})
