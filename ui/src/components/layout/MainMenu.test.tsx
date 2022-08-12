import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'

import * as authentication from '../../helpers/authentication'
import { ADMIN_QUERY } from './Layout'
import { MemoryRouter } from 'react-router-dom'
import MainMenu, { MAPBOX_QUERY } from './MainMenu'

vi.mock('../../helpers/authentication.ts')

const authTokenMock = vi.mocked(authentication.authToken)

afterEach(() => {
  authTokenMock.mockClear()
})

test('Layout sidebar component', async () => {
  authTokenMock.mockImplementation(() => 'test-token')

  const mockedGraphql = [
    {
      request: {
        query: ADMIN_QUERY,
      },
      result: {
        data: {
          myUser: {
            admin: true,
          },
        },
      },
    },
    {
      request: {
        query: MAPBOX_QUERY,
      },
      result: {
        data: {
          mapboxToken: true,
        },
      },
    },
  ]

  render(
    <MockedProvider mocks={mockedGraphql} addTypename={false}>
      <MemoryRouter>
        <MainMenu />
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Timeline')).toBeInTheDocument()
  expect(screen.getByText('Albums')).toBeInTheDocument()

  expect(await screen.findByText('Settings')).toBeInTheDocument()
  expect(await screen.findByText('Places')).toBeInTheDocument()
})
