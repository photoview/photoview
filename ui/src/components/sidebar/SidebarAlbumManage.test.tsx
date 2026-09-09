import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { MemoryRouter } from 'react-router-dom'
import SidebarAlbumManage, {
  MANAGE_ALBUMS_QUERY,
  RENAME_ALBUM_MUTATION,
  MOVE_ALBUM_MUTATION,
  DELETE_ALBUM_MUTATION,
} from './SidebarAlbumManage'

// The delete confirmation uses <Modal>, which relies on @headlessui/react's
// Dialog, which requires IntersectionObserver - not implemented by jsdom.
class MockIntersectionObserver {
  observe() {
    // no-op: not exercised by these tests
  }
  unobserve() {
    // no-op: not exercised by these tests
  }
  disconnect() {
    // no-op: not exercised by these tests
  }
}

beforeAll(() => {
  global.IntersectionObserver =
    MockIntersectionObserver as unknown as typeof IntersectionObserver
})

const albumsQueryMock = {
  request: { query: MANAGE_ALBUMS_QUERY },
  result: {
    data: {
      myAlbums: [
        { id: '1', title: 'Current folder', __typename: 'Album' },
        { id: '2', title: 'Other folder', __typename: 'Album' },
      ],
    },
  },
}

test('excludes the current album from the move destination list', async () => {
  render(
    <MockedProvider mocks={[albumsQueryMock]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumManage albumId="1" albumTitle="Current folder" />
      </MemoryRouter>
    </MockedProvider>
  )

  await waitFor(() => {
    expect(screen.getByText('Other folder')).toBeInTheDocument()
  })
  expect(screen.queryByText('Current folder')).not.toBeInTheDocument()
})

test('renames the album', async () => {
  let requestedVariables: unknown = null

  const renameMock = {
    request: {
      query: RENAME_ALBUM_MUTATION,
      variables: { albumId: '1', newName: 'Renamed folder' },
    },
    result: () => {
      requestedVariables = { albumId: '1', newName: 'Renamed folder' }
      return {
        data: {
          renameAlbum: {
            id: '1',
            title: 'Renamed folder',
            __typename: 'Album',
          },
        },
      }
    },
  }

  render(
    <MockedProvider mocks={[albumsQueryMock, renameMock]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumManage albumId="1" albumTitle="Current folder" />
      </MemoryRouter>
    </MockedProvider>
  )

  const input = screen.getByDisplayValue('Current folder')
  fireEvent.change(input, { target: { value: 'Renamed folder' } })
  fireEvent.click(screen.getByText('Rename'))

  await waitFor(() => {
    expect(requestedVariables).toEqual({
      albumId: '1',
      newName: 'Renamed folder',
    })
  })
})

test('rename button is disabled until the name actually changes', () => {
  render(
    <MockedProvider mocks={[albumsQueryMock]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumManage albumId="1" albumTitle="Current folder" />
      </MemoryRouter>
    </MockedProvider>
  )

  expect(screen.getByText('Rename')).toBeDisabled()

  const input = screen.getByDisplayValue('Current folder')
  fireEvent.change(input, { target: { value: '' } })
  expect(screen.getByText('Rename')).toBeDisabled()
})

test('moves the album to the selected destination', async () => {
  let requestedVariables: unknown = null

  const moveMock = {
    request: {
      query: MOVE_ALBUM_MUTATION,
      variables: { albumId: '1', newParentAlbumId: '2' },
    },
    result: () => {
      requestedVariables = { albumId: '1', newParentAlbumId: '2' }
      return { data: { moveAlbum: { id: '1', __typename: 'Album' } } }
    },
  }

  render(
    <MockedProvider mocks={[albumsQueryMock, moveMock]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumManage albumId="1" albumTitle="Current folder" />
      </MemoryRouter>
    </MockedProvider>
  )

  await waitFor(() => {
    expect(screen.getByText('Other folder')).toBeInTheDocument()
  })

  fireEvent.change(screen.getByRole('combobox'), { target: { value: '2' } })
  fireEvent.click(screen.getByText('Move'))

  await waitFor(() => {
    expect(requestedVariables).toEqual({
      albumId: '1',
      newParentAlbumId: '2',
    })
  })
})

test('deletes the album after confirming', async () => {
  let deleteRequested = false

  const deleteMock = {
    request: {
      query: DELETE_ALBUM_MUTATION,
      variables: { albumId: '1' },
    },
    result: () => {
      deleteRequested = true
      return { data: { deleteAlbum: true } }
    },
  }

  render(
    <MockedProvider mocks={[albumsQueryMock, deleteMock]} addTypename={false}>
      <MemoryRouter>
        <SidebarAlbumManage albumId="1" albumTitle="Current folder" />
      </MemoryRouter>
    </MockedProvider>
  )

  fireEvent.click(screen.getByText('Delete this folder'))

  // Confirmation modal appears with a second "Delete this folder" button
  const confirmButtons = await screen.findAllByText('Delete this folder')
  expect(confirmButtons.length).toBeGreaterThan(1)
  fireEvent.click(confirmButtons[confirmButtons.length - 1])

  await waitFor(() => {
    expect(deleteRequested).toBe(true)
  })
})
