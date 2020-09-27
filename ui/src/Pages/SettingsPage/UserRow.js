import gql from 'graphql-tag'
import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { Mutation } from 'react-apollo'
import {
  Button,
  Checkbox,
  Form,
  Icon,
  Input,
  Modal,
  Table,
} from 'semantic-ui-react'

const updateUserMutation = gql`
  mutation updateUser(
    $id: Int!
    $username: String
    $rootPath: String
    $admin: Boolean
  ) {
    updateUser(
      id: $id
      username: $username
      rootPath: $rootPath
      admin: $admin
    ) {
      id
      username
      rootPath
      admin
    }
  }
`

const deleteUserMutation = gql`
  mutation deleteUser($id: Int!) {
    deleteUser(id: $id) {
      id
      username
    }
  }
`

const changeUserPasswordMutation = gql`
  mutation changeUserPassword($userId: Int!, $password: String!) {
    updateUser(id: $userId, password: $password) {
      id
    }
  }
`

const scanUserMutation = gql`
  mutation scanUser($userId: Int!) {
    scanUser(userId: $userId) {
      success
    }
  }
`

const ChangePasswordModal = ({ onClose, user, ...props }) => {
  const [passwordInput, setPasswordInput] = useState('')

  return (
    <Mutation
      mutation={changeUserPasswordMutation}
      onCompleted={() => {
        onClose && onClose()
      }}
    >
      {changePassword => (
        <Modal {...props}>
          <Modal.Header>Change password</Modal.Header>
          <Modal.Content>
            <p>
              Change password for <b>{user.username}</b>
            </p>
            <Form>
              <Form.Field>
                <label>New password</label>
                <Input
                  placeholder="password"
                  onChange={e => setPasswordInput(e.target.value)}
                  type="password"
                />
              </Form.Field>
            </Form>
          </Modal.Content>
          <Modal.Actions>
            <Button onClick={() => onClose && onClose()}>Cancel</Button>
            <Button
              positive
              onClick={() => {
                changePassword({
                  variables: {
                    userId: user.id,
                    password: passwordInput,
                  },
                })
              }}
            >
              Change password
            </Button>
          </Modal.Actions>
        </Modal>
      )}
    </Mutation>
  )
}

ChangePasswordModal.propTypes = {
  onClose: PropTypes.func,
  user: PropTypes.object.isRequired,
}

const UserRow = ({ user, refetchUsers }) => {
  const [state, setState] = useState({
    ...user,
    editing: false,
  })

  const [showComfirmDelete, setConfirmDelete] = useState(false)
  const [showChangePassword, setChangePassword] = useState(false)

  function updateInput(event, key) {
    setState({
      ...state,
      [key]: event.target.value,
    })
  }

  if (state.editing) {
    return (
      <Mutation
        mutation={updateUserMutation}
        onCompleted={data => {
          setState({
            ...data.updateUser,
            editing: false,
          })
          refetchUsers()
        }}
      >
        {(updateUser, { loading }) => (
          <Table.Row>
            <Table.Cell>
              <Input
                style={{ width: '100%' }}
                placeholder={user.username}
                value={state.username}
                onChange={e => updateInput(e, 'username')}
              />
            </Table.Cell>
            <Table.Cell>
              <Input
                style={{ width: '100%' }}
                placeholder={user.rootPath}
                value={state.rootPath}
                onChange={e => updateInput(e, 'rootPath')}
              />
            </Table.Cell>
            <Table.Cell>
              <Checkbox
                toggle
                checked={state.admin}
                onChange={(_, data) => {
                  setState({
                    ...state,
                    admin: data.checked,
                  })
                }}
              />
            </Table.Cell>
            <Table.Cell>
              <Button.Group>
                <Button
                  negative
                  onClick={() =>
                    setState({
                      ...state.oldState,
                    })
                  }
                >
                  Cancel
                </Button>
                <Button
                  loading={loading}
                  disabled={loading}
                  positive
                  onClick={() =>
                    updateUser({
                      variables: {
                        id: user.id,
                        username: state.username,
                        rootPath: state.rootPath,
                        admin: state.admin,
                      },
                    })
                  }
                >
                  Save
                </Button>
              </Button.Group>
            </Table.Cell>
          </Table.Row>
        )}
      </Mutation>
    )
  }

  return (
    <Mutation
      mutation={deleteUserMutation}
      onCompleted={() => {
        refetchUsers()
      }}
    >
      {deleteUser => (
        <Table.Row>
          <Table.Cell>{user.username}</Table.Cell>
          <Table.Cell>{user.rootPath}</Table.Cell>
          <Table.Cell>
            {user.admin ? <Icon name="checkmark" size="large" /> : null}
          </Table.Cell>
          <Table.Cell>
            <Button.Group>
              <Button
                onClick={() => {
                  setState({ ...state, editing: true, oldState: state })
                }}
              >
                <Icon name="edit" />
                Edit
              </Button>
              <Mutation mutation={scanUserMutation}>
                {(scanUser, { called }) => (
                  <Button
                    disabled={called}
                    onClick={() => scanUser({ variables: { userId: user.id } })}
                  >
                    <Icon name="sync" />
                    Scan
                  </Button>
                )}
              </Mutation>
              <Button onClick={() => setChangePassword(true)}>
                <Icon name="key" />
                Change password
              </Button>
              <ChangePasswordModal
                user={user}
                open={showChangePassword}
                onClose={() => setChangePassword(false)}
              />
              <Button
                negative
                onClick={() => {
                  setConfirmDelete(true)
                }}
              >
                <Icon name="delete" />
                Delete
              </Button>
              <Modal open={showComfirmDelete}>
                <Modal.Header>Delete user</Modal.Header>
                <Modal.Content>
                  <p>
                    {`Are you sure, you want to delete `}
                    <b>{user.username}</b>?
                  </p>
                  <p>{`This action cannot be undone`}</p>
                </Modal.Content>
                <Modal.Actions>
                  <Button onClick={() => setConfirmDelete(false)}>
                    Cancel
                  </Button>
                  <Button
                    negative
                    onClick={() => {
                      setConfirmDelete(false)
                      deleteUser({
                        variables: {
                          id: user.id,
                        },
                      })
                    }}
                  >
                    Delete {user.username}
                  </Button>
                </Modal.Actions>
              </Modal>
            </Button.Group>
          </Table.Cell>
        </Table.Row>
      )}
    </Mutation>
  )
}

UserRow.propTypes = {
  user: PropTypes.object.isRequired,
  refetchUsers: PropTypes.func.isRequired,
}

export default UserRow
