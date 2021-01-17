import React from 'react'
import { Button, Icon, Table, Modal } from 'semantic-ui-react'
import styled from 'styled-components'
import ChangePasswordModal from './UserChangePassword'
import { UserRowProps } from './UserRow'

const PathList = styled.ul`
  margin: 0;
  padding: 0 0 0 12px;
  list-style: none;
`

const ViewUserRow = ({
  user,
  // state,
  setState,
  scanUser,
  deleteUser,
  setChangePassword,
  setConfirmDelete,
  scanUserCalled,
  showChangePassword,
  showConfirmDelete,
}) => {
  const paths = (
    <PathList>
      {user.rootAlbums.map(album => (
        <li key={album.id}>{album.filePath}</li>
      ))}
    </PathList>
  )

  return (
    <Table.Row>
      <Table.Cell>{user.username}</Table.Cell>
      <Table.Cell>{paths}</Table.Cell>
      <Table.Cell>
        {user.admin ? <Icon name="checkmark" size="large" /> : null}
      </Table.Cell>
      <Table.Cell>
        <Button.Group>
          <Button
            onClick={() => {
              setState(state => ({ ...state, editing: true, oldState: state }))
            }}
          >
            <Icon name="edit" />
            Edit
          </Button>
          <Button
            disabled={scanUserCalled}
            onClick={() => scanUser({ variables: { userId: user.id } })}
          >
            <Icon name="sync" />
            Scan
          </Button>
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
          <Modal open={showConfirmDelete}>
            <Modal.Header>Delete user</Modal.Header>
            <Modal.Content>
              <p>
                {`Are you sure, you want to delete `}
                <b>{user.username}</b>?
              </p>
              <p>{`This action cannot be undone`}</p>
            </Modal.Content>
            <Modal.Actions>
              <Button onClick={() => setConfirmDelete(false)}>Cancel</Button>
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
  )
}

ViewUserRow.propTypes = UserRowProps

export default ViewUserRow
