import React, { useState } from 'react'
import PropTypes from 'prop-types'
import { useMutation, useQuery, gql } from '@apollo/client'
import {
  Table,
  Button,
  Dropdown,
  Checkbox,
  Input,
  Icon,
} from 'semantic-ui-react'
import copy from 'copy-to-clipboard'
import { authToken } from '../../authentication'

const sharePhotoQuery = gql`
  query sidbarGetPhotoShares($id: Int!) {
    media(id: $id) {
      id
      shares {
        token
        hasPassword
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
        hasPassword
      }
    }
  }
`

const addPhotoShareMutation = gql`
  mutation sidebarPhotoAddShare($id: Int!, $password: String, $expire: Time) {
    shareMedia(mediaId: $id, password: $password, expire: $expire) {
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

const protectShareMutation = gql`
  mutation sidebarProtectShare($token: String!, $password: String) {
    protectShareToken(token: $token, password: $password) {
      token
      hasPassword
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

const ShareItemMoreDropdown = ({ id, share, isPhoto }) => {
  const query = isPhoto ? sharePhotoQuery : shareAlbumQuery

  const [deleteShare, { loading: deleteShareLoading }] = useMutation(
    deleteShareMutation,
    {
      refetchQueries: [{ query: query, variables: { id } }],
    }
  )

  const [addingPassword, setAddingPassword] = useState(false)
  const showPasswordInput = addingPassword || share.hasPassword

  const [passwordInputValue, setPasswordInputValue] = useState(
    share.hasPassword ? '**********' : ''
  )
  const [passwordHidden, setPasswordHidden] = useState(share.hasPassword)

  const hidePassword = hide => {
    setPasswordHidden(hide)
    if (hide) {
      setPasswordInputValue('**********')
    }
  }

  const [setPassword, { loading: setPasswordLoading }] = useMutation(
    protectShareMutation,
    {
      refetchQueries: [{ query: query, variables: { id } }],
      onCompleted: data => {
        console.log('data', data)
        hidePassword(data.protectShareToken.hasPassword)
      },
      // refetchQueries: [{ query: query, variables: { id } }],
      variables: {
        token: share.token,
      },
    }
  )

  let addPasswordInput = null
  if (showPasswordInput) {
    const setPasswordEvent = event => {
      if (!passwordHidden && passwordInputValue != '' && event.key == 'Enter') {
        event.preventDefault()
        setPassword({
          variables: {
            password: event.target.value,
          },
        })
      }
    }

    addPasswordInput = (
      <Input
        disabled={setPasswordLoading}
        loading={setPasswordLoading}
        style={{ marginTop: 8, marginRight: 0, display: 'block' }}
        onClick={e => e.stopPropagation()}
        value={passwordInputValue}
        type={passwordHidden ? 'password' : 'text'}
        onKeyUp={setPasswordEvent}
        onChange={event => {
          hidePassword(false)
          setPasswordInputValue(event.target.value)
        }}
        placeholder="Password"
        icon={
          <Icon
            name={passwordHidden ? 'lock' : 'arrow right'}
            link={!passwordHidden}
            onClick={setPasswordEvent}
          />
        }
      />
    )
  }

  const checkboxClick = () => {
    const enable = !showPasswordInput
    setAddingPassword(enable)
    if (!enable) {
      setPassword({
        variables: {
          password: null,
        },
      })
      setPasswordInputValue('')
    }
  }

  // const [dropdownOpen, setDropdownOpen] = useState(false)

  return (
    <Dropdown
      // onBlur={event => {
      //   console.log('Blur')
      // }}
      // onClick={() => setDropdownOpen(state => !state)}
      // onClose={() => setDropdownOpen(false)}
      // open={dropdownOpen}
      button
      text="More"
      closeOnChange={false}
      closeOnBlur={false}
    >
      <Dropdown.Menu>
        <Dropdown.Item
          onKeyDown={e => e.stopPropagation()}
          onClick={e => {
            e.stopPropagation()
            checkboxClick()
          }}
        >
          <Checkbox
            label="Password"
            onClick={e => e.stopPropagation()}
            checked={showPasswordInput}
            onChange={() => {
              checkboxClick()
            }}
          />
          {addPasswordInput}
        </Dropdown.Item>
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
  )
}

ShareItemMoreDropdown.propTypes = {
  id: PropTypes.number.isRequired,
  isPhoto: PropTypes.bool.isRequired,
  share: PropTypes.object.isRequired,
}

const SidebarShare = ({ photo, album }) => {
  if ((!photo || !photo.id) && (!album || !album.id)) return null
  if (!authToken()) return null

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

  const [sharePhoto, { loading: sharePhotoLoading }] = useMutation(
    addShareMutation,
    {
      refetchQueries: [{ query: query, variables: { id } }],
    }
  )

  let content = null

  if (sharesError) {
    content = <div>Error: {sharesError.message}</div>
  }

  if (!content && sharesLoading) {
    content = <div>Loading shares...</div>
  }

  if (!content) {
    const shares = isPhoto ? sharesData.media.shares : sharesData.album.shares

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
            <ShareItemMoreDropdown share={share} id={id} isPhoto={isPhoto} />
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
