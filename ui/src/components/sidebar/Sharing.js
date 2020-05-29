import React from 'react'
import PropTypes from 'prop-types'
import { useMutation, useQuery } from 'react-apollo'
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
  mutation sidebareDeleteShare($token: String!) {
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

  const {
    loading: sharesLoading,
    error: sharesError,
    data: sharesData,
  } = useQuery(query, {
    variables: { id },
  })

  const [deleteShare, { loading: deleteShareLoading }] = useMutation(
    deleteShareMutation,
    {
      refetchQueries: [{ query: query, variables: { id } }],
    }
  )

  const [sharePhoto, { loading: sharePhotoLoading }] = useMutation(
    addShareMutation,
    {
      refetchQueries: [{ query: query, variables: { id } }],
    }
  )

  let content = null

  if (!content && sharesError) {
    content = <div>Error: {sharesError.message}</div>
  }

  if (!content && sharesLoading) {
    content = <div>Loading shares...</div>
  }

  if (!content) {
    const shares = isPhoto ? sharesData.photo.shares : sharesData.album.shares

    const optionsRows = shares.map(share => (
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
                <Dropdown.Item
                  text="Delete"
                  icon="delete"
                  disabled={deleteShareLoading}
                  onClick={() => {
                    deleteShare({
                      variables: {
                        token: share.token,
                      },
                    })
                  }}
                />
              </Dropdown.Menu>
            </Dropdown>
          </Button.Group>
        </Table.Cell>
      </Table.Row>
    ))

    if (optionsRows.length == 0) {
      optionsRows.push(
        <Table.Row key="no-shares">
          <Table.Cell colSpan="2">No shares found</Table.Cell>
        </Table.Row>
      )
    }

    content = (
      <div>
        <Table>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell colSpan="2">Public Shares</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>{optionsRows}</Table.Body>
          <Table.Footer>
            <Table.Row>
              <Table.HeaderCell colSpan="2">
                <Button
                  content="Add share"
                  icon="add"
                  floated="right"
                  positive
                  loading={sharePhotoLoading}
                  disabled={sharePhotoLoading}
                  onClick={() => {
                    sharePhoto({
                      variables: {
                        id,
                      },
                    })
                  }}
                />
              </Table.HeaderCell>
            </Table.Row>
          </Table.Footer>
        </Table>
      </div>
    )
  }

  return (
    <div>
      <h2>Sharing options</h2>
      {content}
    </div>
  )
}

SidebarShare.propTypes = {
  photo: PropTypes.object,
  album: PropTypes.object,
}

export default SidebarShare
