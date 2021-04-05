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
import { authToken } from '../../helpers/authentication'
import styled from 'styled-components'
import { useTranslation } from 'react-i18next'

const sharePhotoQuery = gql`
  query sidbarGetPhotoShares($id: ID!) {
    media(id: $id) {
      id
      shares {
        id
        token
        hasPassword
      }
    }
  }
`

const shareAlbumQuery = gql`
  query sidbarGetAlbumShares($id: ID!) {
    album(id: $id) {
      id
      shares {
        id
        token
        hasPassword
      }
    }
  }
`

const addPhotoShareMutation = gql`
  mutation sidebarPhotoAddShare($id: ID!, $password: String, $expire: Time) {
    shareMedia(mediaId: $id, password: $password, expire: $expire) {
      token
    }
  }
`

const addAlbumShareMutation = gql`
  mutation sidebarAlbumAddShare($id: ID!, $password: String, $expire: Time) {
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
  const { t } = useTranslation()
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
            token: share.token,
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
            label={t('login_page.field.password', 'Password')}
            onClick={e => e.stopPropagation()}
            checked={showPasswordInput}
            onChange={() => {
              checkboxClick()
            }}
          />
          {addPasswordInput}
        </Dropdown.Item>
        <Dropdown.Item
          text={t('general.action.delete', 'Delete')}
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
  id: PropTypes.string.isRequired,
  isPhoto: PropTypes.bool.isRequired,
  share: PropTypes.object.isRequired,
}

const ShareButtonGroup = styled(Button.Group)`
  flex-wrap: wrap;
`

const SidebarShare = ({ photo, album }) => {
  const { t } = useTranslation()
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
    content = <div>{t('general.loading.shares', 'Loading shares...')}</div>
  }

  if (!content) {
    const shares = isPhoto ? sharesData.media.shares : sharesData.album.shares

    const optionsRows = shares.map(share => (
      <Table.Row key={share.token}>
        <Table.Cell>
          <b>{t('sidebar.sharing.public_link', 'Public Link')}</b> {share.token}
        </Table.Cell>
        <Table.Cell>
          <ShareButtonGroup>
            <Button
              icon="chain"
              content={t('sidebar.sharing.copy_link', 'Copy Link')}
              onClick={() => {
                copy(`${location.origin}/share/${share.token}`)
              }}
            />
            <ShareItemMoreDropdown share={share} id={id} isPhoto={isPhoto} />
          </ShareButtonGroup>
        </Table.Cell>
      </Table.Row>
    ))

    if (optionsRows.length == 0) {
      optionsRows.push(
        <Table.Row key="no-shares">
          <Table.Cell colSpan="2">
            {t('sidebar.sharing.no_shares_found', 'No shares found')}
          </Table.Cell>
        </Table.Row>
      )
    }

    content = (
      <div>
        <Table>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell colSpan="2">
                {t('sidebar.sharing.table_header', 'Public shares')}
              </Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>{optionsRows}</Table.Body>
          <Table.Footer>
            <Table.Row>
              <Table.HeaderCell colSpan="2">
                <Button
                  content={t('sidebar.sharing.add_share', 'Add shares')}
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
      <h2>{t('sidebar.sharing.title', 'Sharing options')}</h2>
      {content}
    </div>
  )
}

SidebarShare.propTypes = {
  photo: PropTypes.object,
  album: PropTypes.object,
}

export default SidebarShare
