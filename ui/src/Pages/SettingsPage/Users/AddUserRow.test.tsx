import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import AddUserRow, {
  CREATE_USER_MUTATION,
  USER_ADD_ROOT_PATH_MUTATION,
} from './AddUserRow'
import { MockedProvider } from '@apollo/client/testing'
import { ROLE_QUERY } from './RoleSelector'
import { expect } from 'vitest'

vi.mock('react-i18next', () => ({
  /*
   * This is a bit of a hack, but for some reason on the testing library a label of `Please Select` will always
   * assign a value of 1 irrespective of the actual value this DOES NOT happen in the frontend. I think spending the time
   * investigating this is not worth it as it's going to be some weird inner workings of the testing library and hopefully
   * the migration to the latest vitest will fix this issue (When I raise the PR I will add a reference to here so it
   * should get fixed)
   */
  useTranslation: () => {
    const t = (key: string) => key
    return { t }
  },
}))

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
  const usernameInput = screen.getByPlaceholderText('login_page.field.username')
  const pathInput = screen.getByPlaceholderText(
    'login_page.initial_setup.field.photo_path.placeholder'
  )
  const addUserBtn = screen.getByText('settings.users.add_user.submit')
  // Await for role selector to have it's elements loaded from gql
  await screen.findByText('general.please_select')
  const userRoleSelect = screen.getByRole('combobox')
  expect(
    addUserBtn.disabled,
    'User button should be disabled until loaded and role selected'
  ).toBeTruthy()
  console.log('Role select debug')
  screen.debug(userRoleSelect)
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

  const usernameInput = screen.getByPlaceholderText('login_page.field.username')
  const addUserBtn = screen.getByText('settings.users.add_user.submit')
  // Await for role selector to have it's elements loaded from gql
  await screen.findByText('general.please_select')
  const userRoleSelect = screen.getByRole('combobox')

  fireEvent.change(usernameInput, { target: { value: 'testuser' } })
  fireEvent.change(userRoleSelect, { target: { value: '1' } })
  fireEvent.click(addUserBtn)
  await waitFor(() => {
    expect(userAdded).toHaveBeenCalledTimes(1)
  })

  expect(setShow).not.toHaveBeenCalled()
})
