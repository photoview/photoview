import React from 'react'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import { Table, Button, Icon, Dropdown } from 'semantic-ui-react'

const shareQuery = gql`
  query sidbarGetShares($photoId: ID!) {
    photoShares(id: $photoId) {
      token
    }
  }
`

const SidebarShare = ({ photo }) => {
  if (!photo || !photo.id) return null

  return (
    <div>
      <h2>Sharing options</h2>
      <Query query={shareQuery} variables={{ photoId: photo.id }}>
        {({ loading, error, data }) => {
          if (loading) return <div>Loading...</div>
          if (error) return <div>Error: {error}</div>

          const rows = data.photoShares.map(share => (
            <Table.Row key={share.token}>
              <Table.Cell>
                <b>Public Link</b> {share.token}
              </Table.Cell>
              <Table.Cell>
                <Button.Group>
                  <Button icon="chain" content="Copy" />
                  <Dropdown button text="More">
                    <Dropdown.Menu>
                      <Dropdown.Item text="Delete" icon="delete" />
                    </Dropdown.Menu>
                  </Dropdown>
                </Button.Group>
              </Table.Cell>
            </Table.Row>
          ))

          if (rows.length == 0) {
            rows.push(
              <Table.Row>
                <Table.Cell colSpan="2">No shares found</Table.Cell>
              </Table.Row>
            )
          }

          return (
            <div>
              <Table>
                <Table.Header>
                  <Table.Row>
                    <Table.HeaderCell colSpan="2">
                      Public Shares
                    </Table.HeaderCell>
                  </Table.Row>
                </Table.Header>
                <Table.Body>{rows}</Table.Body>
                <Table.Footer>
                  <Table.Row>
                    <Table.HeaderCell colSpan="2">
                      <Button content="New" floated="right" positive />
                    </Table.HeaderCell>
                  </Table.Row>
                </Table.Footer>
              </Table>
            </div>
          )
        }}
      </Query>
    </div>
  )
}

export default SidebarShare
