import React, { useState } from 'react'
import { Mutation } from 'react-apollo'
import { Table, Button, Input, Checkbox } from 'semantic-ui-react'
import gql from 'graphql-tag'

const createUserMutation = gql`
  mutation createUser($username: String, $rootPath: String, $admin: Boolean) {
    createUser(username: $username, rootPath: $rootPath, admin: $admin) {
      id
      username
      rootPath
      admin
    }
  }
`

const initialState = {
  username: '',
  rootPath: '',
  admin: false,
}

const AddUserRow = ({ setShow, show, onUserAdded }) => {
  const [state, setState] = useState(initialState)

  function updateInput(event, key) {
    setState({
      ...state,
      [key]: event.target.value,
    })
  }

  if (!show) {
    return null
  }

  return (
    <Mutation
      mutation={createUserMutation}
      onCompleted={data => {
        onUserAdded()
      }}
    >
      {(createUser, { loading, data }) => (
        <Table.Row>
          <Table.Cell>
            <Input
              placeholder="Username"
              value={state.username}
              onChange={e => updateInput(e, 'username')}
            />
          </Table.Cell>
          <Table.Cell>
            <Input
              placeholder="/path/to/photos"
              value={state.rootPath}
              onChange={e => updateInput(e, 'rootPath')}
            />
          </Table.Cell>
          <Table.Cell>
            <Checkbox
              toggle
              checked={state.admin}
              onChange={(e, data) => {
                console.log(data)
                setState({
                  ...state,
                  admin: data.checked,
                })
              }}
            />
          </Table.Cell>
          <Table.Cell>
            <Button.Group>
              <Button negative onClick={e => setShow(false)}>
                Cancel
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
                      rootPath: state.rootPath,
                      admin: state.admin,
                    },
                  })
                  setState(initialState)
                }}
              >
                Add User
              </Button>
            </Button.Group>
          </Table.Cell>
        </Table.Row>
      )}
    </Mutation>
  )
}

export default AddUserRow
