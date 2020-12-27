import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Button, Checkbox, Icon, Input, Table } from 'semantic-ui-react'
import { UserRowProps } from './UserRow'

const RootPathListItem = styled.li`
  display: flex;
  justify-content: space-between;
  align-items: center;
`

const EditRootPath = ({ filePath, removePath }) => (
  <RootPathListItem>
    <span>{filePath}</span>
    <Button negative onClick={() => removePath()}>
      <Icon name="remove" />
      Remove
    </Button>
  </RootPathListItem>
)

EditRootPath.propTypes = {
  filePath: PropTypes.string.isRequired,
  removePath: PropTypes.func.isRequired,
}

const NewRootPathInput = styled(Input)`
  width: 100%;
  margin-top: 24px;
`

const EditNewRootPath = ({ state, updateInput }) => (
  <li>
    <NewRootPathInput
      style={{ width: '100%' }}
      value={state.rootPath}
      onChange={e => updateInput(e, 'rootPath')}
      action={{
        positive: true,
        icon: 'add',
        content: 'Add',
      }}
    />
  </li>
)

EditNewRootPath.propTypes = {
  state: PropTypes.object.isRequired,
  updateInput: PropTypes.func.isRequired,
}

const RootPathList = styled.ul`
  margin: 0;
  padding: 0;
  list-style: none;
`

const EditRootPaths = ({ user, state, updateInput }) => {
  const editRows = user.rootAlbums.map(album => (
    <EditRootPath
      key={album.id}
      filePath={album.filePath}
      removePath={() => {}}
    />
  ))

  return (
    <RootPathList>
      {editRows}
      <EditNewRootPath state={state} updateInput={updateInput} />
    </RootPathList>
  )
}

EditRootPaths.propTypes = {
  updateInput: PropTypes.func.isRequired,
  user: PropTypes.object.isRequired,
  state: PropTypes.object.isRequired,
}

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
        <EditRootPaths user={user} state={state} updateInput={updateInput} />
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
