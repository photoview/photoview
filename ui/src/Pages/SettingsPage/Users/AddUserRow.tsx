import { gql, useMutation } from '@apollo/client'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button, Checkbox, Input, Table } from 'semantic-ui-react'

const CREATE_USER_MUTATION = gql`
  mutation createUser($username: String!, $admin: Boolean!) {
    createUser(username: $username, admin: $admin) {
      id
      username
      admin
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
  admin: false,
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

  const [addRootPath, { loading: addRootPathLoading }] = useMutation(
    USER_ADD_ROOT_PATH_MUTATION,
    {
      onCompleted: () => {
        setState(initialState)
        onUserAdded()
      },
      onError: () => {
        setState(initialState)
        onUserAdded()
      },
    }
  )

  const [createUser, { loading: createUserLoading }] = useMutation(
    CREATE_USER_MUTATION,
    {
      onCompleted: ({ createUser: { id } }) => {
        if (state.rootPath) {
          addRootPath({
            variables: {
              id: id,
              rootPath: state.rootPath,
            },
          })
        } else {
          setState(initialState)
        }
      },
    }
  )

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
    <Table.Row>
      <Table.Cell>
        <Input
          placeholder={t('login_page.field.username', 'Username')}
          value={state.username}
          onChange={e => updateInput(e, 'username')}
        />
      </Table.Cell>
      <Table.Cell>
        <Input
          placeholder={t(
            'login_page.initial_setup.field.photo_path.placeholder',
            '/path/to/photos'
          )}
          value={state.rootPath}
          onChange={e => updateInput(e, 'rootPath')}
        />
      </Table.Cell>
      <Table.Cell>
        <Checkbox
          toggle
          checked={state.admin}
          onChange={(e, data) => {
            setState({
              ...state,
              admin: data.checked || false,
            })
          }}
        />
      </Table.Cell>
      <Table.Cell>
        <Button.Group>
          <Button negative onClick={() => setShow(false)}>
            {t('general.action.cancel', 'Cancel')}
          </Button>
          <Button
            type="submit"
            loading={loading}
            disabled={loading}
            positive
            onClick={() => {
              createUser({
                variables: {
                  username: state.username,
                  admin: state.admin,
                },
              })
            }}
          >
            {t('settings.users.add_user.submit', 'Add user')}
          </Button>
        </Button.Group>
      </Table.Cell>
    </Table.Row>
  )
}

export default AddUserRow
