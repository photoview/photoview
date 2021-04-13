import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { Button, Icon, Input } from 'semantic-ui-react'
import styled from 'styled-components'
import { USERS_QUERY } from './UsersTable'
import { useTranslation } from 'react-i18next'
import { USER_ADD_ROOT_PATH_MUTATION } from './AddUserRow'
import {
  userRemoveAlbumPathMutation,
  userRemoveAlbumPathMutationVariables,
} from './__generated__/userRemoveAlbumPathMutation'
import {
  settingsUsersQuery_user,
  settingsUsersQuery_user_rootAlbums,
} from './__generated__/settingsUsersQuery'
import { userAddRootPath } from './__generated__/userAddRootPath'

const USER_REMOVE_ALBUM_PATH_MUTATION = gql`
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

type EditRootPathProps = {
  album: settingsUsersQuery_user_rootAlbums
  user: settingsUsersQuery_user
}

const EditRootPath = ({ album, user }: EditRootPathProps) => {
  const { t } = useTranslation()
  const [removeAlbumPath, { loading }] = useMutation<
    userRemoveAlbumPathMutation,
    userRemoveAlbumPathMutationVariables
  >(USER_REMOVE_ALBUM_PATH_MUTATION, {
    refetchQueries: [
      {
        query: USERS_QUERY,
      },
    ],
  })

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

const NewRootPathInput = styled(Input)`
  width: 100%;
  margin-top: 24px;
`

type EditNewRootPathProps = {
  userID: string
}

const EditNewRootPath = ({ userID }: EditNewRootPathProps) => {
  const { t } = useTranslation()
  const [value, setValue] = useState('')
  const [addRootPath, { loading }] = useMutation<userAddRootPath>(
    USER_ADD_ROOT_PATH_MUTATION,
    {
      refetchQueries: [
        {
          query: USERS_QUERY,
        },
      ],
    }
  )

  return (
    <li>
      <NewRootPathInput
        style={{ width: '100%' }}
        value={value}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
          setValue(e.target.value)
        }
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

const RootPathList = styled.ul`
  margin: 0;
  padding: 0;
  list-style: none;
`

type EditRootPathsProps = {
  user: settingsUsersQuery_user
}

export const EditRootPaths = ({ user }: EditRootPathsProps) => {
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
