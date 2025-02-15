import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import AlbumPage from './AlbumPage'
import { renderWithProviders } from '../../helpers/testUtils'

vi.mock('../../hooks/useScrollPagination')

test('AlbumPage renders', () => {
  renderWithProviders(<AlbumPage />, {
    mocks: [],
    initialEntries: ['/album/1'],
    path: "/album/:id",
    route: <AlbumPage />
  })

  expect(screen.getByText('Sort')).toBeInTheDocument()
  expect(screen.getByLabelText('Sort direction')).toBeInTheDocument()

  screen.debug()
})
