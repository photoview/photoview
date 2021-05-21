import React from 'react'
import { useTranslation } from 'react-i18next'
import { Button, Checkbox, Input, Table } from 'semantic-ui-react'
import { EditRootPaths } from './EditUserRowRootPaths'
import { UserRowChildProps } from './UserRow'

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
              admin: data.checked || false,
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
                ...state,
                ...state.oldState,
              }))
            }
          >
            {t('general.action.cancel', 'Cancel')}
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
                  admin: state.admin,
                },
              })
            }
          >
            {t('general.action.save', 'Save')}
          </Button>
        </Button.Group>
      </Table.Cell>
    </Table.Row>
  )
}

export default EditUserRow
