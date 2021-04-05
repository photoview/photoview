import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { Button, Icon, Input } from 'semantic-ui-react'
import styled from 'styled-components'
import { USERS_QUERY } from './UsersTable'
import { useTranslation } from 'react-i18next'

const userAddRootPathMutation = gql`
  mutation userAddRootPath($id: ID!, $rootPath: String!) {
    userAddRootPath(id: $id, rootPath: $rootPath) {
      id
    }
  }
`

const userRemoveAlbumPathMutation = gql`
  mutation userRemoveAlbumPathMutation($userId: ID!, $albumId: ID!) {
    userRemoveRootAlbum(userId: $userId, albumId: $albumId) {
      id
    }
  }
`

const RootPathListItem = styled.li`
  display: flex;
  justify-content: space-between;
  align-items: center;
`

const EditRootPath = ({ album, user }) => {
  const { t } = useTranslation()
  const [removeAlbumPath, { loading }] = useMutation(
    userRemoveAlbumPathMutation,
    {
      refetchQueries: [
        {
          query: USERS_QUERY,
        },
      ],
    }
  )

  return (
    <RootPathListItem>
      <span>{album.filePath}</span>
      <Button
        negative
        disabled={loading}
        onClick={() =>
          removeAlbumPath({
            variables: {
              userId: user.id,
              albumId: album.id,
            },
          })
        }
      >
        <Icon name="remove" />
        {t('general.action.remove', 'Remove')}
      </Button>
    </RootPathListItem>
  )
}

EditRootPath.propTypes = {
  album: PropTypes.object.isRequired,
  user: PropTypes.object.isRequired,
}

const NewRootPathInput = styled(Input)`
  width: 100%;
  margin-top: 24px;
`

const EditNewRootPath = ({ userID }) => {
  const { t } = useTranslation()
  const [value, setValue] = useState('')
  const [addRootPath, { loading }] = useMutation(userAddRootPathMutation, {
    refetchQueries: [
      {
        query: USERS_QUERY,
      },
    ],
  })

  return (
    <li>
      <NewRootPathInput
        style={{ width: '100%' }}
        value={value}
        onChange={e => setValue(e.target.value)}
        disabled={loading}
        action={{
          positive: true,
          icon: 'add',
          content: t('general.action.add', 'Add'),
          onClick: () => {
            setValue('')
            addRootPath({
              variables: {
                id: userID,
                rootPath: value,
              },
            })
          },
        }}
      />
    </li>
  )
}

EditNewRootPath.propTypes = {
  userID: PropTypes.string.isRequired,
}

const RootPathList = styled.ul`
  margin: 0;
  padding: 0;
  list-style: none;
`

export const EditRootPaths = ({ user }) => {
  const editRows = user.rootAlbums.map(album => (
    <EditRootPath key={album.id} album={album} user={user} />
  ))

  return (
    <RootPathList>
      {editRows}
      <EditNewRootPath userID={user.id} />
    </RootPathList>
  )
}

EditRootPaths.propTypes = {
  user: PropTypes.object.isRequired,
}
