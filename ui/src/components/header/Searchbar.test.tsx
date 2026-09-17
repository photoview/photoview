import React from 'react'
import { InMemoryCache } from '@apollo/client'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import * as authentication from '../../helpers/authentication'
import SearchBar, {
  SEARCH_QUERY,
  SEARCHBAR_USER_PREFERENCES_QUERY,
} from './Searchbar'

vi.mock('../../helpers/authentication.ts')

vi.mocked(authentication.authToken).mockReturnValue('test-token')

const preferencesMock = (searchResultLimit: number | null) => ({
  request: { query: SEARCHBAR_USER_PREFERENCES_QUERY },
  result: {
    data: {
      myUserPreferences: {
        id: '1',
        searchResultLimit,
        __typename: 'UserPreferences',
      },
    },
  },
})

// What the dropdown actually asks for: an unset or unlimited preference is
// clamped to the number of rows it will render at most.
const DROPDOWN_MAX_ROWS = 500

const searchVariables = (limit: number) => ({
  query: 'vac',
  limitMedia: limit,
  limitAlbums: limit,
})

const albumResults = (count: number) =>
  Array.from({ length: count }, (_, i) => ({
    id: `album-${i}`,
    title: `Vacation ${i}`,
    thumbnail: {
      thumbnail: { url: `/thumb-${i}.jpg`, __typename: 'MediaURL' },
      __typename: 'Media',
    },
    __typename: 'Album',
  }))

const renderSearchBar = (mocks: readonly unknown[], cache?: InMemoryCache) => {
  render(
    <MockedProvider mocks={mocks as never} cache={cache} addTypename={false}>
      <MemoryRouter>
        <SearchBar />
      </MemoryRouter>
    </MockedProvider>
  )
}

test('passes the saved result limit to the search', async () => {
  let sawLimit = false

  renderSearchBar([
    preferencesMock(25),
    {
      request: { query: SEARCH_QUERY, variables: searchVariables(25) },
      result: () => {
        sawLimit = true
        return {
          data: {
            search: {
              query: 'vac',
              albums: [],
              media: [],
              __typename: 'SearchResult',
            },
          },
        }
      },
    },
  ])

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')
  await waitFor(() => expect(sawLimit).toBe(true), { timeout: 15000 })
}, 20000)

test('re-issues the search when the limit preference changes to unlimited', async () => {
  let sawLimited = false
  let sawUnlimited = false
  // Must match MockedProvider's addTypename={false}, or the cache rewrites
  // the outgoing documents and no mock matches anymore.
  const cache = new InMemoryCache({ addTypename: false })

  const emptyResult = {
    data: {
      search: {
        query: 'vac',
        albums: [],
        media: [],
        __typename: 'SearchResult',
      },
    },
  }

  renderSearchBar(
    [
      preferencesMock(10),
      {
        request: { query: SEARCH_QUERY, variables: searchVariables(10) },
        result: () => {
          sawLimited = true
          return emptyResult
        },
      },
      {
        request: {
          query: SEARCH_QUERY,
          variables: searchVariables(DROPDOWN_MAX_ROWS),
        },
        result: () => {
          sawUnlimited = true
          return emptyResult
        },
      },
    ],
    cache
  )

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')
  await waitFor(() => expect(sawLimited).toBe(true), { timeout: 15000 })

  // Saving "0" (unlimited) in the settings while the searchbar stays
  // mounted: a defined-to-defined change, which a plain "has it resolved
  // yet" check would miss entirely.
  cache.writeQuery({
    query: SEARCHBAR_USER_PREFERENCES_QUERY,
    data: {
      myUserPreferences: {
        id: '1',
        searchResultLimit: 0,
        __typename: 'UserPreferences',
      },
    },
  })

  await waitFor(() => expect(sawUnlimited).toBe(true), { timeout: 15000 })
}, 40000)

test('drops the thumbnails once the result list gets long', async () => {
  renderSearchBar([
    preferencesMock(null),
    {
      request: {
        query: SEARCH_QUERY,
        variables: searchVariables(DROPDOWN_MAX_ROWS),
      },
      result: {
        data: {
          search: {
            query: 'vac',
            albums: albumResults(60),
            media: [],
            __typename: 'SearchResult',
          },
        },
      },
    },
  ])

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')

  await waitFor(
    () =>
      expect(
        screen.getByRole('list', { name: 'albums' }).children
      ).toHaveLength(60),
    { timeout: 15000 }
  )

  expect(
    screen.queryAllByRole('img'),
    'a list this long renders as compact text rows'
  ).toHaveLength(0)
}, 20000)

test('keeps the thumbnails for a short result list', async () => {
  renderSearchBar([
    preferencesMock(null),
    {
      request: {
        query: SEARCH_QUERY,
        variables: searchVariables(DROPDOWN_MAX_ROWS),
      },
      result: {
        data: {
          search: {
            query: 'vac',
            albums: albumResults(3),
            media: [],
            __typename: 'SearchResult',
          },
        },
      },
    },
  ])

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')

  await waitFor(() => expect(screen.queryAllByRole('img')).toHaveLength(3), {
    timeout: 15000,
  })
}, 20000)
