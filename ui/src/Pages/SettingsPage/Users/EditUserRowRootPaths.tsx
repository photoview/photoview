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
  userUpdateRootPathLevel,
  userUpdateRootPathLevelVariables,
} from './__generated__/userUpdateRootPathLevel'
import {
  settingsUsersQuery_user,
  settingsUsersQuery_user_rootAlbums,
} from './__generated__/settingsUsersQuery'
import { userAddRootPath } from './__generated__/userAddRootPath'
import { Button, TextField } from '../../../primitives/form/Input'
import Dropdown from '../../../primitives/form/Dropdown'
import { AlbumPermissionLevel } from '../../../__generated__/globalTypes'
import { albumPermissionLevelOptions } from '../../../helpers/albumPermissions'

const USER_REMOVE_ALBUM_PATH_MUTATION = gql`
  mutation userRemoveAlbumPathMutation($userId: ID!, $albumId: ID!) {
    userRemoveRootAlbum(userId: $userId, albumId: $albumId) {
      id
    }
  }
`

const USER_UPDATE_ROOT_PATH_LEVEL_MUTATION = gql`
  mutation userUpdateRootPathLevel(
    $id: ID!
    $albumId: ID!
    $level: AlbumPermissionLevel!
  ) {
    userUpdateRootPathLevel(id: $id, albumId: $albumId, level: $level) {
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
  const levelOptions = albumPermissionLevelOptions(t)
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

  const [updateLevel, { loading: updateLevelLoading }] = useMutation<
    userUpdateRootPathLevel,
    userUpdateRootPathLevelVariables
  >(USER_UPDATE_ROOT_PATH_LEVEL_MUTATION, {
    refetchQueries: [
      {
        query: USERS_QUERY,
      },
    ],
    // The global Apollo error link already shows a toast; this only
    // consumes the rejected promise so it isn't left unhandled.
    onError: () => undefined,
  })

  const currentLevel =
    album.permissions?.find(p => p.user.id === user.id)?.level ??
    AlbumPermissionLevel.READ

  return (
    <li className="flex justify-between items-center gap-2">
      <span>{album.filePath}</span>
      <div className="flex gap-1">
        <Dropdown
          items={levelOptions}
          selected={currentLevel}
          disabled={updateLevelLoading}
          setSelected={value =>
            updateLevel({
              variables: {
                id: user.id,
                albumId: album.id,
                level: value as AlbumPermissionLevel,
              },
            })
          }
        />
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
      </div>
    </li>
  )
}

type EditNewRootPathProps = {
  userID: string
}

const EditNewRootPath = ({ userID }: EditNewRootPathProps) => {
  const { t } = useTranslation()
  const levelOptions = albumPermissionLevelOptions(t)
  const [value, setValue] = useState('')
  const [level, setLevel] = useState<AlbumPermissionLevel>(
    AlbumPermissionLevel.READ
  )
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
      <Dropdown
        items={levelOptions}
        selected={level}
        disabled={loading}
        setSelected={value => setLevel(value as AlbumPermissionLevel)}
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
              level,
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
