import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
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
import { Button, TextField } from '../../../primitives/form/Input'

const USER_REMOVE_ALBUM_PATH_MUTATION = gql`
  mutation userRemoveAlbumPathMutation($userId: ID!, $albumId: ID!) {
    userRemoveRootAlbum(userId: $userId, albumId: $albumId) {
      id
    }
  }
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
    <li className="flex justify-between">
      <span>{album.filePath}</span>
      <Button
        variant="negative"
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
        {t('general.action.remove', 'Remove')}
      </Button>
    </li>
  )
}

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
    <li className="flex gap-1 mt-2">
      <TextField
        value={value}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
          setValue(e.target.value)
        }
        disabled={loading}
      />
      <Button
        variant="positive"
        disabled={loading}
        onClick={() => {
          setValue('')
          addRootPath({
            variables: {
              id: userID,
              rootPath: value,
            },
          })
        }}
      >
        {t('general.action.add', 'Add')}
      </Button>
    </li>
  )
}

type EditRootPathsProps = {
  user: settingsUsersQuery_user
}

export const EditRootPaths = ({ user }: EditRootPathsProps) => {
  const editRows = user.rootAlbums.map(album => (
    <EditRootPath key={album.id} album={album} user={user} />
  ))

  return (
    <ul>
      {editRows}
      <EditNewRootPath userID={user.id} />
    </ul>
  )
}
