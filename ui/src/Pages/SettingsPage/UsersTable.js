import React, { useState } from 'react'

import { Table, Loader, Button, Icon } from 'semantic-ui-react'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import UserRow from './UserRow'
import AddUserRow from './AddUserRow'

const usersQuery = gql`
  query settingsUsersQuery {
    user {
      id
      username
      rootPath
      admin
    }
  }
`

const UsersTable = () => {
  const [showAddUser, setShowAddUser] = useState(false)

  return (
    <Query query={usersQuery}>
      {({ loading, error, data, refetch }) => {
        let userRows = []
        if (data && data.user) {
          userRows = data.user.map(user => (
            <UserRow user={user} refetchUsers={refetch} key={user.id} />
          ))
        }

        return (
          <div style={{ marginTop: 24 }}>
            <h2>Users</h2>
            <Loader active={loading} />
            <Table celled>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>Username</Table.HeaderCell>
                  <Table.HeaderCell>Photo path</Table.HeaderCell>
                  <Table.HeaderCell>Admin</Table.HeaderCell>
                  <Table.HeaderCell>Action</Table.HeaderCell>
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
                      onClick={e => setShowAddUser(true)}
                    >
                      <Icon name="add" />
                      New user
                    </Button>
                  </Table.HeaderCell>
                </Table.Row>
              </Table.Footer>
            </Table>
          </div>
        )
      }}
    </Query>
  )
}

export default UsersTable
