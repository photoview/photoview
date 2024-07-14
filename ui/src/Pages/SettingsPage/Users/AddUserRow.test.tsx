import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import AddUserRow, {
  CREATE_USER_MUTATION,
  USER_ADD_ROOT_PATH_MUTATION,
} from './AddUserRow'
import { MockedProvider } from '@apollo/client/testing'
import { ROLE_QUERY } from './RoleSelector'
import { expect } from 'vitest'

const gqlMock = [
  {
    request: {
      query: ROLE_QUERY,
    },
    result: {
      data: {
        roles: [
          { id: '1', name: 'ADMIN' },
          { id: '2', name: 'USER' },
          { id: '3', name: 'DEMO' },
        ],
      },
    },
  },
  {
    request: {
      query: CREATE_USER_MUTATION,
      variables: { username: 'testuser', roleId: '1' },
    },
    result: {
      data: {
        createUser: {
          id: '123',
          username: 'testuser',
          admin: false,
          role: {
            id: '123',
            name: 'ADMIN',
          },
          __typename: 'User',
        },
      },
    },
  },
  {
    request: {
      query: USER_ADD_ROOT_PATH_MUTATION,
      variables: { id: '123', rootPath: '/tmp' },
    },
    result: { data: { userAddRootPath: { id: '567', __typename: 'Album' } } },
  },
]

test('Add user with username and path', async () => {
  const userAdded = vi.fn()
  const setShow = vi.fn()

  render(
    <MockedProvider addTypename={true} mocks={gqlMock}>
      <table>
        <tbody>
          <AddUserRow onUserAdded={userAdded} setShow={setShow} show={true} />
        </tbody>
      </table>
    </MockedProvider>
  )
  const usernameInput = screen.getByPlaceholderText('Username')
  const pathInput = screen.getByPlaceholderText('/path/to/photos')
  const addUserBtn = screen.getByText('Add user') 
  const userRoleSelect = await screen.findByPlaceholderText('Please Select')

  expect(
    addUserBtn.disabled,
    'User button should be disabled until loaded and role selected'
  ).toBeTruthy()
  fireEvent.change(userRoleSelect, { target: { value: '1' } })
  expect(
    addUserBtn.disabled,
    'User button should be enabled once role has been selected'
  ).toBeFalsy()
  fireEvent.change(usernameInput, { target: { value: 'testuser' } })
  fireEvent.change(pathInput, { target: { value: '/tmp' } })
  fireEvent.click(addUserBtn)

  await waitFor(() => {
    expect(userAdded).toHaveBeenCalledTimes(1)
  })

  expect(setShow).not.toHaveBeenCalled()
})

test('Add user with only username', async () => {
  const userAdded = vi.fn()
  const setShow = vi.fn()

  render(
    <MockedProvider addTypename={true} mocks={gqlMock}>
      <table>
        <tbody>
          <AddUserRow onUserAdded={userAdded} setShow={setShow} show={true} />
        </tbody>
      </table>
    </MockedProvider>
  )

  const usernameInput = screen.getByPlaceholderText('Username')
  const addUserBtn = screen.getByText('Add user')
  const userRoleSelect = await screen.findByPlaceholderText('Please Select')

  fireEvent.change(usernameInput, { target: { value: 'testuser' } })
  fireEvent.change(userRoleSelect, { target: { value: '1' } })
  fireEvent.click(addUserBtn)
  await waitFor(() => {
    expect(userAdded).toHaveBeenCalledTimes(1)
  })

  expect(setShow).not.toHaveBeenCalled()
})
