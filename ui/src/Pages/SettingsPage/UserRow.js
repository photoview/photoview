import React, { useState } from 'react'
import { Mutation } from 'react-apollo'
import { Table, Icon, Button, Input, Checkbox, Modal } from 'semantic-ui-react'
import gql from 'graphql-tag'

const updateUserMutation = gql`
  mutation updateUser(
    $id: ID!
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
  mutation deleteUser($id: ID!) {
    deleteUser(id: $id) {
      id
      username
    }
  }
`

const UserRow = ({ user, refetchUsers }) => {
  const [state, setState] = useState({
    ...user,
    editing: false,
  })

  const [showComfirmDelete, setConfirmDelete] = useState(false)

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
        {(updateUser, { loading, data }) => (
          <Table.Row>
            <Table.Cell>
              <Input
                placeholder={user.username}
                value={state.username}
                onChange={e => updateInput(e, 'username')}
              />
            </Table.Cell>
            <Table.Cell>
              <Input
                placeholder={user.rootPath}
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
                    admin: data.checked,
                  })
                }}
              />
            </Table.Cell>
            <Table.Cell>
              <Button.Group>
                <Button
                  negative
                  onClick={e =>
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
      {(deleteUser, { loading, data }) => (
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

export default UserRow
