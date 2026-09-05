import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { MemoryRouter } from 'react-router-dom'
import SidebarAlbumNewFolder, {
  CREATE_ALBUM_FOLDER_MUTATION,
} from './SidebarAlbumNewFolder'

test('creates a folder with the entered name', async () => {
  let requestedVariables: unknown = null

  const mocks = [
    {
      request: {
        query: CREATE_ALBUM_FOLDER_MUTATION,
        variables: { parentAlbumId: '1', name: 'Vacation' },
      },
      result: () => {
        requestedVariables = { parentAlbumId: '1', name: 'Vacation' }
        return {
          data: {
            createAlbumFolder: { id: '2', title: 'Vacation', __typename: 'Album' },
          },
        }
      },
    },
  ]

  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumNewFolder albumId="1" />
      </MemoryRouter>
    </MockedProvider>
  )

  const input = screen.getByPlaceholderText('Folder name')
  fireEvent.change(input, { target: { value: 'Vacation' } })
  fireEvent.click(screen.getByText('Create'))

  await waitFor(() => {
    expect(requestedVariables).toEqual({
      parentAlbumId: '1',
      name: 'Vacation',
    })
  })

  // Input clears after a successful creation
  await waitFor(() => {
    expect(screen.getByPlaceholderText('Folder name')).toHaveValue('')
  })
})

test('create button is disabled until a name is entered', () => {
  render(
    <MockedProvider mocks={[]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumNewFolder albumId="1" />
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Create')).toBeDisabled()

  fireEvent.change(screen.getByPlaceholderText('Folder name'), {
    target: { value: 'Vacation' },
  })

  expect(screen.getByText('Create')).not.toBeDisabled()
})
