import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'

import * as authentication from '../../helpers/authentication'
import { ADMIN_QUERY } from './Layout'
import { MemoryRouter } from 'react-router-dom'
import MainMenu, { SITE_INFO_FEATURE_FLAGS_QUERY } from './MainMenu'

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
        query: SITE_INFO_FEATURE_FLAGS_QUERY,
      },
      result: {
        data: {
          siteInfo: {
            faceDetectionEnabled: true,
            mapEnabled: true,
          },
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

  expect(await screen.findByText('Places')).toBeInTheDocument()
  expect(screen.getByText('Settings')).toBeInTheDocument()
})

test('hides Places when map is disabled', async () => {
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
        query: SITE_INFO_FEATURE_FLAGS_QUERY,
      },
      result: {
        data: {
          siteInfo: {
            faceDetectionEnabled: true,
            mapEnabled: false,
          },
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

  // Wait for the query to resolve by checking People appears (faceDetectionEnabled: true)
  expect(await screen.findByText('People')).toBeInTheDocument()

  expect(screen.getByText('Timeline')).toBeInTheDocument()
  expect(screen.getByText('Albums')).toBeInTheDocument()
  expect(screen.queryByText('Places')).not.toBeInTheDocument()
  expect(screen.getByText('Settings')).toBeInTheDocument()
})
