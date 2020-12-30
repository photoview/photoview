import React from 'react'
import { Button, Checkbox, Input, Table } from 'semantic-ui-react'
import { EditRootPaths } from './EditUserRowRootPaths'
import { UserRowProps } from './UserRow'

const EditUserRow = ({
  user,
  state,
  setState,
  updateUser,
  updateUserLoading,
}) => {
  function updateInput(event, key) {
    setState(state => ({
      ...state,
      [key]: event.target.value,
    }))
  }

  return (
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
        <EditRootPaths user={user} />
      </Table.Cell>
      <Table.Cell>
        <Checkbox
          toggle
          checked={state.admin}
          onChange={(_, data) => {
            setState(state => ({
              ...state,
              admin: data.checked,
            }))
          }}
        />
      </Table.Cell>
      <Table.Cell>
        <Button.Group>
          <Button
            negative
            onClick={() =>
              setState(state => ({
                ...state.oldState,
              }))
            }
          >
            Cancel
          </Button>
          <Button
            loading={updateUserLoading}
            disabled={updateUserLoading}
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
  )
}

EditUserRow.propTypes = UserRowProps

export default EditUserRow
