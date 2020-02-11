import React from 'react'
import PropTypes from 'prop-types'
import { Query, Mutation } from 'react-apollo'
import gql from 'graphql-tag'
import { Table, Button, Dropdown } from 'semantic-ui-react'
import copy from 'copy-to-clipboard'

const sharePhotoQuery = gql`
  query sidbarGetPhotoShares($id: Int!) {
    photo(id: $id) {
      id
      shares {
        token
      }
    }
  }
`

const shareAlbumQuery = gql`
  query sidbarGetAlbumShares($id: Int!) {
    album(id: $id) {
      id
      shares {
        token
      }
    }
  }
`

const addPhotoShareMutation = gql`
  mutation sidebarPhotoAddShare($id: Int!, $password: String, $expire: Time) {
    sharePhoto(photoId: $id, password: $password, expire: $expire) {
      token
    }
  }
`

const addAlbumShareMutation = gql`
  mutation sidebarAlbumAddShare($id: Int!, $password: String, $expire: Time) {
    shareAlbum(albumId: $id, password: $password, expire: $expire) {
      token
    }
  }
`

const deleteShareMutation = gql`
  mutation sidebareDeleteShare($token: ID!) {
    deleteShareToken(token: $token) {
      token
    }
  }
`

const SidebarShare = ({ photo, album }) => {
  if ((!photo || !photo.id) && (!album || !album.id)) return null
  if (!localStorage.getItem('token')) return null

  const isPhoto = !!photo
  const id = isPhoto ? photo.id : album.id

  const query = isPhoto ? sharePhotoQuery : shareAlbumQuery
  const addShareMutation = isPhoto
    ? addPhotoShareMutation
    : addAlbumShareMutation

  return (
    <div>
      <h2>Sharing options</h2>
      <Query query={query} variables={{ id }}>
        {({ loading, error, data, refetch }) => {
          if (loading) return <div>Loading...</div>
          if (error) return <div>Error: {error.message}</div>

          let shares = isPhoto ? data.photo.shares : data.album.shares

          const rows = shares.map(share => (
            <Table.Row key={share.token}>
              <Table.Cell>
                <b>Public Link</b> {share.token}
              </Table.Cell>
              <Table.Cell>
                <Button.Group>
                  <Button
                    icon="chain"
                    content="Copy link"
                    onClick={() => {
                      copy(`${location.origin}/share/${share.token}`)
                    }}
                  />
                  <Dropdown button text="More">
                    <Dropdown.Menu>
                      <Mutation
                        mutation={deleteShareMutation}
                        onCompleted={() => {
                          refetch()
                        }}
                      >
                        {(deleteShare, { loading, error, data }) => {
                          return (
                            <Dropdown.Item
                              text="Delete"
                              icon="delete"
                              disabled={loading}
                              onClick={() => {
                                deleteShare({
                                  variables: {
                                    token: share.token,
                                  },
                                })
                              }}
                            />
                          )
                        }}
                      </Mutation>
                    </Dropdown.Menu>
                  </Dropdown>
                </Button.Group>
              </Table.Cell>
            </Table.Row>
          ))

          if (rows.length == 0) {
            rows.push(
              <Table.Row key="no-shares">
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
                      <Mutation
                        mutation={addShareMutation}
                        onCompleted={() => {
                          refetch()
                        }}
                      >
                        {(sharePhoto, { loading, error, data }) => {
                          return (
                            <Button
                              content="Add share"
                              icon="add"
                              floated="right"
                              positive
                              loading={loading}
                              disabled={loading}
                              onClick={() => {
                                sharePhoto({
                                  variables: {
                                    id,
                                  },
                                })
                              }}
                            />
                          )
                        }}
                      </Mutation>
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

SidebarShare.propTypes = {
  photo: PropTypes.object,
  album: PropTypes.object,
}

export default SidebarShare
