import { gql, useMutation } from '@apollo/client'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { TextField, Button, ButtonGroup } from '../../../primitives/form/Input'
import { TableRow, TableCell } from '../../../primitives/Table'
import { createUser, createUserVariables } from './__generated__/createUser'
import {
  userAddRootPath,
  userAddRootPathVariables,
} from './__generated__/userAddRootPath'
import { RoleSelector } from './RoleSelector'

export const CREATE_USER_MUTATION = gql`
  mutation createUser($username: String!, $roleId: ID!) {
    createUser(username: $username, roleId: $roleId) {
      id
      username
      role {
        id
        name
      }
      __typename
    }
  }
`

export const USER_ADD_ROOT_PATH_MUTATION = gql`
  mutation userAddRootPath($id: ID!, $rootPath: String!) {
    userAddRootPath(id: $id, rootPath: $rootPath) {
      id
    }
  }
`

const initialState = {
  username: '',
  rootPath: '',
  roleId: '',
  userAdded: false,
}

type AddUserRowProps = {
  setShow: React.Dispatch<React.SetStateAction<boolean>>
  show: boolean
  onUserAdded(): void
}

const AddUserRow = ({ setShow, show, onUserAdded }: AddUserRowProps) => {
  const { t } = useTranslation()
  const [state, setState] = useState(initialState)

  const finished = () => {
    setState(initialState)
    onUserAdded()
  }

  const [addRootPath, { loading: addRootPathLoading }] = useMutation<
    userAddRootPath,
    userAddRootPathVariables
  >(USER_ADD_ROOT_PATH_MUTATION, {
    onCompleted: () => {
      finished()
    },
    onError: () => {
      finished()
    },
  })

  const [createUser, { loading: createUserLoading }] = useMutation<
    createUser,
    createUserVariables
  >(CREATE_USER_MUTATION, {
    onCompleted: ({ createUser: { id } }) => {
      if (state.rootPath) {
        addRootPath({
          variables: {
            id: id,
            rootPath: state.rootPath,
          },
        })
      } else {
        finished()
      }
    },
  })

  const loading = addRootPathLoading || createUserLoading

  function updateInput(
    event: React.ChangeEvent<HTMLInputElement>,
    key: string
  ) {
    setState({
      ...state,
      [key]: event.target.value,
    })
  }

  if (!show) {
    return null
  }

  return (
    <TableRow>
      <TableCell>
        <TextField
          placeholder={t('login_page.field.username', 'Username')}
          value={state.username}
          onChange={e => updateInput(e, 'username')}
        />
      </TableCell>
      <TableCell>
        <TextField
          placeholder={t(
            'login_page.initial_setup.field.photo_path.placeholder',
            '/path/to/photos'
          )}
          value={state.rootPath}
          onChange={e => updateInput(e, 'rootPath')}
        />
      </TableCell>
      <TableCell>
        <RoleSelector
          onRoleSelected={i => {
            setState(state => ({
              ...state,
              roleId: i,
            }))
          }}
          selected={state.roleId}
        />
      </TableCell>
      <TableCell>
        <ButtonGroup>
          <Button variant="negative" onClick={() => setShow(false)}>
            {t('general.action.cancel', 'Cancel')}
          </Button>
          <Button
            type="submit"
            disabled={loading}
            variant="positive"
            onClick={() => {
              createUser({
                variables: {
                  username: state.username,
                  roleId: state.roleId,
                },
              })
            }}
          >
            {t('settings.users.add_user.submit', 'Add user')}
          </Button>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  )
}

export default AddUserRow
