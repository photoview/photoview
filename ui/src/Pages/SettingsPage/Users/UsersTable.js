import React, { useState } from 'react'

import { Table, Loader, Button, Icon } from 'semantic-ui-react'
import { useQuery, gql } from '@apollo/client'
import UserRow from './UserRow'
import AddUserRow from './AddUserRow'
import { SectionTitle } from '../SettingsPage'
import { useTranslation } from 'react-i18next'

export const USERS_QUERY = gql`
  query settingsUsersQuery {
    user {
      id
      username
      # rootPath
      admin
      rootAlbums {
        id
        filePath
      }
    }
  }
`

const UsersTable = () => {
  const { t } = useTranslation()
  const [showAddUser, setShowAddUser] = useState(false)

  const { loading, error, data, refetch } = useQuery(USERS_QUERY)

  if (error) {
    return `Users table error: ${error.message}`
  }

  let userRows = []
  if (data && data.user) {
    userRows = data.user.map(user => (
      <UserRow user={user} refetchUsers={refetch} key={user.id} />
    ))
  }

  return (
    <div>
      <SectionTitle>{t('settings.users.title', 'Users')}</SectionTitle>
      <Loader active={loading} />
      <Table celled>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>
              {t('settings.users.table.column_names.username', 'Username')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('settings.users.table.column_names.photo_path', 'Photo path')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('settings.users.table.column_names.admin', 'Admin')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('settings.users.table.column_names.action', 'Action')}
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {userRows}
          <AddUserRow
            show={showAddUser}
            setShow={setShowAddUser}
            onUserAdded={() => {
              setShowAddUser(false)
              refetch()
            }}
          />
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan="4">
              <Button
                positive
                disabled={showAddUser}
                floated="right"
                onClick={() => setShowAddUser(true)}
              >
                <Icon name="add" />
                {t('settings.users.table.new_user', 'New user')}
              </Button>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </div>
  )
}

export default UsersTable
