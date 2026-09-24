import React, { useState } from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { InMemoryCache } from '@apollo/client'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom'
import SearchBar from './Searchbar'
import { AlbumTreeSearchContext } from '../albumTree/AlbumTreeSearchContext'

const NavigateButton = () => {
  const navigate = useNavigate()

  return <button onClick={() => navigate('/album/1')}>go</button>
}

// Mirrors AlbumTreeSearchProvider, but exposes the current query so the test
// can assert on what the tree would be filtering by.
const Harness = () => {
  const [query, setQuery] = useState('')

  return (
    <MockedProvider
      mocks={[]}
      cache={new InMemoryCache({ addTypename: false })}
    >
      <MemoryRouter initialEntries={['/']}>
        <AlbumTreeSearchContext.Provider value={{ query, setQuery }}>
          <SearchBar />
          <NavigateButton />
          <Routes>
            <Route
              path="*"
              element={<span data-testid="tree-query">{query}</span>}
            />
          </Routes>
        </AlbumTreeSearchContext.Provider>
      </MemoryRouter>
    </MockedProvider>
  )
}

test('navigating away clears the filter shared with the album tree', async () => {
  render(<Harness />)

  await userEvent.type(screen.getByRole('searchbox'), 'ausflug')
  expect(screen.getByTestId('tree-query')).toHaveTextContent('ausflug')

  await userEvent.click(screen.getByRole('button', { name: 'go' }))

  // Leaving the input filtered by a term it no longer shows would strand the
  // tree on a search the user cannot see or undo.
  expect(screen.getByTestId('tree-query')).toBeEmptyDOMElement()
  expect(screen.getByRole('searchbox')).toHaveValue('')
})
