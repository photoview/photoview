import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import MapStyleSettings, {
  MAP_STYLES_QUERY,
  SET_MAP_STYLE_LIGHT_MUTATION,
  SET_MAP_STYLE_DARK_MUTATION,
} from './MapStyleSettings'

test('load MapStyleSettings with defaults — shows checkbox unchecked, hides URL fields', async () => {
  const graphqlMocks = [
    {
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: null,
            mapStyleDark: null,
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
  expect(screen.getByText('Customize map style')).toBeInTheDocument()

  await waitFor(() => {
    expect(screen.getByRole('checkbox')).not.toBeChecked()
  })

  expect(screen.queryByText('Light mode style URL')).not.toBeInTheDocument()
  expect(screen.queryByText('Dark mode style URL')).not.toBeInTheDocument()
})

test('shows URL fields when customize checkbox is checked', async () => {
  const graphqlMocks = [
    {
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: null,
            mapStyleDark: null,
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

  await waitFor(() => {
    expect(screen.getByRole('checkbox')).not.toBeChecked()
  })

  fireEvent.click(screen.getByRole('checkbox'))

  expect(screen.getByText('Light mode style URL')).toBeInTheDocument()
  expect(screen.getByText('Dark mode style URL')).toBeInTheDocument()
})

test('loads with custom URLs — checkbox is checked and URL fields visible', async () => {
  const graphqlMocks = [
    {
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://example.com/custom-light',
            mapStyleDark: 'https://example.com/custom-dark',
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

  await waitFor(() => {
    expect(screen.getByRole('checkbox')).toBeChecked()
  })

  expect(screen.getByDisplayValue('https://example.com/custom-light')).toBeInTheDocument()
  expect(screen.getByDisplayValue('https://example.com/custom-dark')).toBeInTheDocument()
})

test('update light style URL on blur', async () => {
  const setLightMock = vi.fn()
  const graphqlMocks = [
    {
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://example.com/old-light',
            mapStyleDark: 'https://example.com/custom-dark',
          },
        },
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_LIGHT_MUTATION,
        variables: { url: 'https://example.com/custom-light' },
      },
      result: () => {
        setLightMock()
        return {
          data: { setMapStyleLight: 'https://example.com/custom-light' },
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
    expect(screen.getByDisplayValue('https://example.com/old-light')).toBeInTheDocument()
  })

  const lightInput = screen.getByDisplayValue('https://example.com/old-light')
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
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://example.com/custom-light',
            mapStyleDark: 'https://example.com/old-dark',
          },
        },
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_DARK_MUTATION,
        variables: { url: 'https://example.com/custom-dark' },
      },
      result: () => {
        setDarkMock()
        return {
          data: { setMapStyleDark: 'https://example.com/custom-dark' },
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
    expect(screen.getByDisplayValue('https://example.com/old-dark')).toBeInTheDocument()
  })

  const darkInput = screen.getByDisplayValue('https://example.com/old-dark')
  fireEvent.change(darkInput, { target: { value: 'https://example.com/custom-dark' } })
  fireEvent.keyDown(darkInput, { key: 'Enter' })

  await waitFor(() => {
    expect(setDarkMock).toHaveBeenCalled()
  })
})

test('unchecking customize resets URLs to null', async () => {
  const setLightMock = vi.fn()
  const setDarkMock = vi.fn()
  const graphqlMocks = [
    {
      request: { query: MAP_STYLES_QUERY },
      result: {
        data: {
          siteInfo: {
            mapStyleLight: 'https://example.com/custom-light',
            mapStyleDark: 'https://example.com/custom-dark',
          },
        },
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_LIGHT_MUTATION,
        variables: { url: null },
      },
      result: () => {
        setLightMock()
        return { data: { setMapStyleLight: null } }
      },
    },
    {
      request: {
        query: SET_MAP_STYLE_DARK_MUTATION,
        variables: { url: null },
      },
      result: () => {
        setDarkMock()
        return { data: { setMapStyleDark: null } }
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
    expect(screen.getByRole('checkbox')).toBeChecked()
  })

  fireEvent.click(screen.getByRole('checkbox'))

  await waitFor(() => {
    expect(setLightMock).toHaveBeenCalled()
    expect(setDarkMock).toHaveBeenCalled()
  })

  expect(screen.queryByText('Light mode style URL')).not.toBeInTheDocument()
  expect(screen.queryByText('Dark mode style URL')).not.toBeInTheDocument()
})
