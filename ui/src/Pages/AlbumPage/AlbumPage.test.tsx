import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import AlbumPage from './AlbumPage'

vi.mock('../../hooks/useScrollPagination')

test('AlbumPage renders', () => {
  render(
    <MockedProvider mocks={[]}>
      <MemoryRouter initialEntries={['/album/1']}>
        <Routes>
          <Route path="/album/:id" element={<AlbumPage />} />
        </Routes>
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Sort')).toBeInTheDocument()
  expect(screen.getByLabelText('Sort direction')).toBeInTheDocument()

  screen.debug()
})
