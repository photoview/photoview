import React from 'react'
import { useTranslation } from 'react-i18next'
import { EditRootPaths } from './EditUserRowRootPaths'
import { UserRowChildProps } from './UserRow'
import { TableRow, TableCell } from '../../../primitives/Table'
import { TextField } from '../../../primitives/form/Input'
import { Button, ButtonGroup } from '../../../primitives/form/Input'
import { RoleSelector } from './RoleSelector'

const EditUserRow = ({
  user,
  state,
  setState,
  updateUser,
  updateUserLoading,
}: UserRowChildProps) => {
  const { t } = useTranslation()

  function updateInput(
    event: React.ChangeEvent<HTMLInputElement>,
    key: string
  ) {
    setState(state => ({
      ...state,
      [key]: event.target.value,
    }))
  }

  return (
    <TableRow>
      <TableCell>
        <TextField
          style={{ width: '100%' }}
          placeholder={user.username}
          value={state.username}
          onChange={e => updateInput(e, 'username')}
        />
      </TableCell>
      <TableCell>
        <EditRootPaths user={user} />
      </TableCell>
      <TableCell>
        <RoleSelector
          onRoleSelected={i => {
            setState(state => ({
              ...state,
              role: {
                ...state.role,
                id: i,
              },
            }))
          }}
          selected={state.role.id}
        />
      </TableCell>
      <TableCell>
        <ButtonGroup>
          <Button
            variant="negative"
            onClick={() =>
              setState(state => ({
                ...state,
                ...state.oldState,
              }))
            }
          >
            {t('general.action.cancel', 'Cancel')}
          </Button>
          <Button
            disabled={updateUserLoading}
            variant="positive"
            onClick={() =>
              updateUser({
                variables: {
                  id: user.id,
                  username: state.username,
                  roleId: state.role.id,
                },
              })
            }
          >
            {t('general.action.save', 'Save')}
          </Button>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  )
}

export default EditUserRow
