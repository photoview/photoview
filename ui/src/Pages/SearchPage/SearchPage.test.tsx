import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom'
import SearchPage, { SEARCH_PAGE_QUERY } from './SearchPage'

vi.mock('../../components/layout/Layout', () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
}))

const searchPageMock = (query: string, albumTitle: string) => ({
  request: { query: SEARCH_PAGE_QUERY, variables: { query, limit: 500 } },
  result: {
    data: {
      search: {
        albums: [
          {
            id: `album-${albumTitle}`,
            title: albumTitle,
            thumbnail: null,
            __typename: 'Album',
          },
        ],
        media: [],
        __typename: 'SearchResult',
      },
    },
  },
})

const SearchAgainButton = () => {
  const navigate = useNavigate()
  return (
    <button onClick={() => navigate('/search?q=second')}>search again</button>
  )
}

test('searching again from the results page queries the new term', async () => {
  render(
    <MockedProvider
      mocks={[
        searchPageMock('first', 'First album'),
        searchPageMock('second', 'Second album'),
      ]}
      addTypename={false}
    >
      <MemoryRouter initialEntries={['/search?q=first']}>
        <SearchAgainButton />
        <Routes>
          <Route path="/search" element={<SearchPage />} />
        </Routes>
      </MemoryRouter>
    </MockedProvider>
  )

  await waitFor(() => expect(screen.getByText('First album')).toBeVisible())

  // The regression: the page read the URL once on mount, so this second
  // search kept showing the first one's results.
  await userEvent.click(screen.getByText('search again'))

  await waitFor(() => expect(screen.getByText('Second album')).toBeVisible())
})
