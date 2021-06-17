import React from 'react'
import { Trans, useTranslation } from 'react-i18next'
import { Icon, Modal } from 'semantic-ui-react'
import Checkbox from '../../../primitives/form/Checkbox'
import { Button } from '../../../primitives/form/Input'
import { TableCell, TableRow } from '../../../primitives/Table'
import ChangePasswordModal from './UserChangePassword'
import { UserRowChildProps } from './UserRow'

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
    <ul>
      {user.rootAlbums.map(album => (
        <li key={album.id}>{album.filePath}</li>
      ))}
    </ul>
  )

  return (
    <TableRow>
      <TableCell>{user.username}</TableCell>
      <TableCell>{paths}</TableCell>
      <TableCell>
        <Checkbox label="Admin" disabled checked={user.admin} />
      </TableCell>
      <TableCell>
        <div className="flex gap-1">
          <Button
            onClick={() => {
              setState(state => {
                const oldState = { ...state }
                delete oldState.oldState
                return { ...state, editing: true, oldState }
              })
            }}
          >
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
            variant="negative"
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
                variant="negative"
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
        </div>
      </TableCell>
    </TableRow>
  )
}

export default ViewUserRow
