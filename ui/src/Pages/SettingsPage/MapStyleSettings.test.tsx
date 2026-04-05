import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import MapStyleSettings, {
  MAP_STYLES_QUERY,
  SET_MAP_STYLE_LIGHT_MUTATION,
  SET_MAP_STYLE_DARK_MUTATION,
} from './MapStyleSettings'

test('load MapStyleSettings with defaults', async () => {
  const graphqlMocks = [
    {
      request: {
        query: MAP_STYLES_QUERY,
      },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://tiles.openfreemap.org/styles/positron',
            mapStyleDark: 'https://tiles.openfreemap.org/styles/dark',
          },
        },
      },
    },
  ]

  render(
    <MockedProvider
      mocks={graphqlMocks}
      addTypename={false}
      defaultOptions={{
        watchQuery: { fetchPolicy: 'no-cache' },
        query: { fetchPolicy: 'no-cache' },
      }}
    >
      <MapStyleSettings />
    </MockedProvider>
  )

  expect(screen.getByText('Map Styles')).toBeInTheDocument()
  expect(screen.getByText('Light mode style URL')).toBeInTheDocument()
  expect(screen.getByText('Dark mode style URL')).toBeInTheDocument()

  await waitFor(() => {
    expect(screen.getByDisplayValue('https://tiles.openfreemap.org/styles/positron')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://tiles.openfreemap.org/styles/dark')).toBeInTheDocument()
  })
})

test('update light style URL on blur', async () => {
  const setLightMock = vi.fn()
  const graphqlMocks = [
    {
      request: {
        query: MAP_STYLES_QUERY,
      },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://tiles.openfreemap.org/styles/positron',
            mapStyleDark: 'https://tiles.openfreemap.org/styles/dark',
          },
        },
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_LIGHT_MUTATION,
        variables: {
          url: 'https://example.com/custom-light',
        },
      },
      result: () => {
        setLightMock()
        return {
          data: {
            setMapStyleLight: 'https://example.com/custom-light',
          },
        }
      },
    },
  ]

  render(
    <MockedProvider
      mocks={graphqlMocks}
      addTypename={false}
      defaultOptions={{
        watchQuery: { fetchPolicy: 'no-cache' },
        query: { fetchPolicy: 'no-cache' },
      }}
    >
      <MapStyleSettings />
    </MockedProvider>
  )

  await waitFor(() => {
    expect(screen.getByDisplayValue('https://tiles.openfreemap.org/styles/positron')).toBeInTheDocument()
  })

  const lightInput = screen.getByDisplayValue('https://tiles.openfreemap.org/styles/positron')
  fireEvent.change(lightInput, { target: { value: 'https://example.com/custom-light' } })
  fireEvent.blur(lightInput)

  await waitFor(() => {
    expect(setLightMock).toHaveBeenCalled()
  })
})

test('update dark style URL on Enter', async () => {
  const setDarkMock = vi.fn()
  const graphqlMocks = [
    {
      request: {
        query: MAP_STYLES_QUERY,
      },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://tiles.openfreemap.org/styles/positron',
            mapStyleDark: 'https://tiles.openfreemap.org/styles/dark',
          },
        },
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_DARK_MUTATION,
        variables: {
          url: 'https://example.com/custom-dark',
        },
      },
      result: () => {
        setDarkMock()
        return {
          data: {
            setMapStyleDark: 'https://example.com/custom-dark',
          },
        }
      },
    },
  ]

  render(
    <MockedProvider
      mocks={graphqlMocks}
      addTypename={false}
      defaultOptions={{
        watchQuery: { fetchPolicy: 'no-cache' },
        query: { fetchPolicy: 'no-cache' },
      }}
    >
      <MapStyleSettings />
    </MockedProvider>
  )

  await waitFor(() => {
    expect(screen.getByDisplayValue('https://tiles.openfreemap.org/styles/dark')).toBeInTheDocument()
  })

  const darkInput = screen.getByDisplayValue('https://tiles.openfreemap.org/styles/dark')
  fireEvent.change(darkInput, { target: { value: 'https://example.com/custom-dark' } })
  fireEvent.keyDown(darkInput, { key: 'Enter' })

  await waitFor(() => {
    expect(setDarkMock).toHaveBeenCalled()
  })
})
