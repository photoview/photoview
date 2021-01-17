import { gql, useMutation } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { Button, Checkbox, Input, Table } from 'semantic-ui-react'

const createUserMutation = gql`
  mutation createUser(
    $username: String!
    $rootPath: String!
    $admin: Boolean!
  ) {
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

  const [createUser, { loading }] = useMutation(createUserMutation, {
    onCompleted: () => {
      onUserAdded()
    },
  })

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
            setState({
              ...state,
              admin: data.checked,
            })
          }}
        />
      </Table.Cell>
      <Table.Cell>
        <Button.Group>
          <Button negative onClick={() => setShow(false)}>
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
  )
}

AddUserRow.propTypes = {
  setShow: PropTypes.func.isRequired,
  show: PropTypes.bool.isRequired,
  onUserAdded: PropTypes.func.isRequired,
}

export default AddUserRow
