import React from 'react'
import { render, screen } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import SidebarAlbumUpload from './SidebarAlbumUpload'

test('renders a file picker and a folder picker', () => {
  render(
    <MockedProvider mocks={[]} addTypename={false}>
      <SidebarAlbumUpload albumId="1" />
    </MockedProvider>
  )

  expect(screen.getByText('Select files')).toBeInTheDocument()
  expect(screen.getByText('Select folder')).toBeInTheDocument()
})
