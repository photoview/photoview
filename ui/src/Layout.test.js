import '@testing-library/jest-dom'

import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'

import Layout, { ADMIN_QUERY, MAPBOX_QUERY, SideMenu } from './Layout'
import { MemoryRouter } from 'react-router-dom'

import * as authentication from './helpers/authentication'

jest.mock('./helpers/authentication.js')

test('Layout component', async () => {
  render(
    <Layout>
      <p>layout_content</p>
    </Layout>
  )

  expect(screen.getByTestId('Layout')).toBeInTheDocument()
  expect(screen.getByText('layout_content')).toBeInTheDocument()
})

afterEach(() => {
  authentication.authToken.mockClear()
})

test('Layout sidebar component', async () => {
  authentication.authToken.mockImplementation(() => true)

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
        <SideMenu />
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Photos')).toBeInTheDocument()
  expect(screen.getByText('Albums')).toBeInTheDocument()

  expect(await screen.findByText('Settings')).toBeInTheDocument()
  expect(await screen.findByText('Places')).toBeInTheDocument()
})
