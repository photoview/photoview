import React, { useState } from 'react'
import {
  Table,
  TableHeader,
  TableHeaderCell,
  TableRow,
  TableBody,
  TableFooter,
  TableScrollWrapper,
} from '../../../primitives/Table'
import { useQuery, gql } from '@apollo/client'
import UserRow from './UserRow'
import AddUserRow from './AddUserRow'
import { SectionTitle } from '../SettingsPage'
import { useTranslation } from 'react-i18next'
import { settingsUsersQuery } from './__generated__/settingsUsersQuery'
import { Button } from '../../../primitives/form/Input'
import Loader from '../../../primitives/Loader'

export const USERS_QUERY = gql`
  query settingsUsersQuery {
    user {
      id
      username
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

  const { loading, error, data, refetch } =
    useQuery<settingsUsersQuery>(USERS_QUERY)

  if (error) {
    return <div>{`Users table error: ${error.message}`}</div>
  }

  let userRows: JSX.Element[] = []
  if (data?.user) {
    userRows = data.user.map(user => (
      <UserRow user={user} refetchUsers={refetch} key={user.id} />
    ))
  }

  return (
    <div>
      <SectionTitle>{t('settings.users.title', 'Users')}</SectionTitle>
      <Loader active={loading} />
      <TableScrollWrapper>
        <Table className="w-full max-w-6xl">
          <TableHeader>
            <TableRow>
              <TableHeaderCell>
                {t('settings.users.table.column_names.username', 'Username')}
              </TableHeaderCell>
              <TableHeaderCell>
                {t(
                  'settings.users.table.column_names.photo_path',
                  'Photo path'
                )}
              </TableHeaderCell>
              <TableHeaderCell>
                {t(
                  'settings.users.table.column_names.capabilities',
                  'Capabilities'
                )}
              </TableHeaderCell>
              <TableHeaderCell className="w-0 whitespace-nowrap">
                {t('settings.users.table.column_names.action', 'Action')}
              </TableHeaderCell>
            </TableRow>
          </TableHeader>

          <TableBody>
            {userRows}
            <AddUserRow
              show={showAddUser}
              setShow={setShowAddUser}
              onUserAdded={() => {
                setShowAddUser(false)
                refetch()
              }}
            />
          </TableBody>

          <TableFooter>
            <TableRow>
              <TableHeaderCell colSpan={4} className="text-right">
                <Button
                  variant="positive"
                  background="white"
                  disabled={showAddUser}
                  onClick={() => setShowAddUser(true)}
                >
                  {t('settings.users.table.new_user', 'New user')}
                </Button>
              </TableHeaderCell>
            </TableRow>
          </TableFooter>
        </Table>
      </TableScrollWrapper>
    </div>
  )
}

export default UsersTable
