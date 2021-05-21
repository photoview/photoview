import React from 'react'
import { Trans, useTranslation } from 'react-i18next'
import { Button, Icon, Table, Modal } from 'semantic-ui-react'
import styled from 'styled-components'
import ChangePasswordModal from './UserChangePassword'
import { UserRowChildProps } from './UserRow'

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
}: UserRowChildProps) => {
  const { t } = useTranslation()
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
              setState(state => {
                const oldState = { ...state }
                delete oldState.oldState
                return { ...state, editing: true, oldState }
              })
            }}
          >
            <Icon name="edit" />
            {t('settings.users.table.row.action.edit', 'Edit')}
          </Button>
          <Button
            disabled={scanUserCalled}
            onClick={() => scanUser({ variables: { userId: user.id } })}
          >
            <Icon name="sync" />
            {t('settings.users.table.row.action.scan', 'Scan')}
          </Button>
          <Button onClick={() => setChangePassword(true)}>
            <Icon name="key" />
            {t(
              'settings.users.table.row.action.change_password',
              'Change password'
            )}
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
            {t('settings.users.table.row.action.delete', 'Delete')}
          </Button>
          <Modal open={showConfirmDelete}>
            <Modal.Header>
              {t('settings.users.confirm_delete_user.title', 'Delete user')}
            </Modal.Header>
            <Modal.Content>
              <Trans
                t={t}
                i18nKey="settings.users.confirm_delete_user.description"
              >
                <p>
                  {`Are you sure, you want to delete `}
                  <b>{user.username}</b>?
                </p>
                <p>{`This action cannot be undone`}</p>
              </Trans>
            </Modal.Content>
            <Modal.Actions>
              <Button onClick={() => setConfirmDelete(false)}>
                {t('general.action.cancel', 'Cancel')}
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
                {t(
                  'settings.users.confirm_delete_user.action',
                  'Delete {{user}}',
                  { user: user.username }
                )}
              </Button>
            </Modal.Actions>
          </Modal>
        </Button.Group>
      </Table.Cell>
    </Table.Row>
  )
}

export default ViewUserRow
