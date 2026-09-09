import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import * as authentication from '../../helpers/authentication'
import SearchBar, {
  SEARCH_QUERY,
  SEARCHBAR_USER_PREFERENCES_QUERY,
} from './Searchbar'
import { SHOW_HIDDEN_ALBUMS_PREFERENCE_QUERY } from '../../hooks/useShowHiddenAlbums'

vi.mock('../../helpers/authentication.ts')

vi.mocked(authentication.authToken).mockReturnValue('test-token')
vi.mocked(authentication.readStoredSearchQuery).mockReturnValue('')
vi.mocked(authentication.writeStoredSearchQuery).mockImplementation(vi.fn())

const emptySearchResult = {
  search: { query: 'vac', albums: [], media: [], __typename: 'SearchResult' },
}

test('re-issues the search once showHiddenAlbums resolves after the first debounced fetch', async () => {
  let sawShowHiddenTrue = false

  const mocks = [
    {
      request: { query: SEARCHBAR_USER_PREFERENCES_QUERY },
      result: {
        data: {
          myUserPreferences: {
            id: '1',
            searchResultLimit: null,
            __typename: 'UserPreferences',
          },
        },
      },
    },
    {
      // Comfortably past the 250ms debounce so the first search has
      // already fired (with showHidden: false) by the time this resolves.
      request: { query: SHOW_HIDDEN_ALBUMS_PREFERENCE_QUERY },
      delay: 1000,
      result: {
        data: {
          myUserPreferences: {
            id: '1',
            showHiddenAlbums: true,
            __typename: 'UserPreferences',
          },
        },
      },
    },
    {
      request: {
        query: SEARCH_QUERY,
        variables: {
          query: 'vac',
          limitMedia: undefined,
          limitAlbums: undefined,
          showHidden: false,
        },
      },
      result: { data: emptySearchResult },
    },
    {
      request: {
        query: SEARCH_QUERY,
        variables: {
          query: 'vac',
          limitMedia: undefined,
          limitAlbums: undefined,
          showHidden: true,
        },
      },
      result: () => {
        sawShowHiddenTrue = true
        return { data: emptySearchResult }
      },
    },
  ]

  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        <SearchBar />
      </MemoryRouter>
    </MockedProvider>
  )

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')

  // The regression: without the fix, this never happens and the test
  // times out here instead. A generous timeout since this suite runs on
  // modest hardware where a busy parallel test run can push real
  // setTimeout callbacks well past their nominal delay.
  await waitFor(() => expect(sawShowHiddenTrue).toBe(true), {
    timeout: 15000,
  })
}, 20000)

test('caps the dropdown to 5 rows even when the search returns more', async () => {
  const media = Array.from({ length: 7 }, (_, i) => ({
    id: `media-${i}`,
    title: `Photo ${i}`,
    thumbnail: { url: `/thumb-${i}.jpg`, __typename: 'MediaURL' },
    album: { id: 'album-1', __typename: 'Album' },
    __typename: 'Media',
  }))

  const mocks = [
    {
      request: { query: SEARCHBAR_USER_PREFERENCES_QUERY },
      result: {
        data: {
          myUserPreferences: {
            id: '1',
            searchResultLimit: null,
            __typename: 'UserPreferences',
          },
        },
      },
    },
    {
      request: { query: SHOW_HIDDEN_ALBUMS_PREFERENCE_QUERY },
      result: {
        data: {
          myUserPreferences: {
            id: '1',
            showHiddenAlbums: false,
            __typename: 'UserPreferences',
          },
        },
      },
    },
    {
      request: {
        query: SEARCH_QUERY,
        variables: {
          query: 'vac',
          limitMedia: undefined,
          limitAlbums: undefined,
          showHidden: false,
        },
      },
      result: {
        data: {
          search: { query: 'vac', albums: [], media, __typename: 'SearchResult' },
        },
      },
    },
  ]

  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        <SearchBar />
      </MemoryRouter>
    </MockedProvider>
  )

  await userEvent.type(screen.getByPlaceholderText('Search'), 'vac')

  await waitFor(() =>
    expect(screen.getByRole('list', { name: 'media' }).children).toHaveLength(
      5
    )
  )
})
