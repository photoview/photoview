import React, { useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import { Button } from '../../primitives/form/Input'
import Dropdown from '../../primitives/form/Dropdown'
import { AlbumPermissionLevel } from '../../__generated__/globalTypes'
import { albumPermissionLevelOptions } from '../../helpers/albumPermissions'
import { albumPermissionsQuery } from './__generated__/albumPermissionsQuery'
import {
  grantAlbumAccess,
  grantAlbumAccessVariables,
} from './__generated__/grantAlbumAccess'
import {
  revokeAlbumAccess,
  revokeAlbumAccessVariables,
} from './__generated__/revokeAlbumAccess'

export const ALBUM_PERMISSIONS_QUERY = gql`
  query albumPermissionsQuery($albumId: ID!) {
    shareableUsers {
      id
      username
    }
    album(id: $albumId) {
      id
      permissions {
        user {
          id
          username
        }
        level
      }
    }
  }
`

export const GRANT_ALBUM_ACCESS_MUTATION = gql`
  mutation grantAlbumAccess(
    $albumId: ID!
    $userId: ID!
    $level: AlbumPermissionLevel!
  ) {
    grantAlbumAccess(albumId: $albumId, userId: $userId, level: $level) {
      user {
        id
        username
      }
      level
    }
  }
`

export const REVOKE_ALBUM_ACCESS_MUTATION = gql`
  mutation revokeAlbumAccess($albumId: ID!, $userId: ID!) {
    revokeAlbumAccess(albumId: $albumId, userId: $userId)
  }
`

type SidebarAlbumSharingProps = {
  albumId: string
}

const SidebarAlbumSharing = ({ albumId }: SidebarAlbumSharingProps) => {
  const { t } = useTranslation()
  const levelOptions = albumPermissionLevelOptions(t)

  const [newUserId, setNewUserId] = useState('')
  const [newLevel, setNewLevel] = useState<AlbumPermissionLevel>(
    AlbumPermissionLevel.READ
  )

  const { data, refetch } = useQuery<albumPermissionsQuery>(
    ALBUM_PERMISSIONS_QUERY,
    { variables: { albumId } }
  )

  const [grantAccess, { loading: granting, error: grantError }] = useMutation<
    grantAlbumAccess,
    grantAlbumAccessVariables
  >(GRANT_ALBUM_ACCESS_MUTATION, {
    onCompleted: () => {
      setNewUserId('')
      refetch()
    },
    // Without this, a rejected mutation is an unhandled promise rejection -
    // grantError above already renders it.
    onError: () => undefined,
  })

  const [revokeAccess, { error: revokeError }] = useMutation<
    revokeAlbumAccess,
    revokeAlbumAccessVariables
  >(REVOKE_ALBUM_ACCESS_MUTATION, {
    onCompleted: () => refetch(),
    // Without this, a rejected mutation is an unhandled promise rejection,
    // and revokeError above would never be set.
    onError: () => undefined,
  })

  const permissions = data?.album.permissions ?? []
  const shareableUsers = (data?.shareableUsers ?? []).filter(
    user => !permissions.some(p => p.user.id === user.id)
  )

  const userOptions = shareableUsers.map(user => ({
    value: user.id,
    label: user.username,
  }))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.sharing.title', 'Shared with')}
      </SidebarSectionTitle>
      <div className="mx-4">
        {permissions.map(permission => (
          <div
            key={permission.user.id}
            className="flex justify-between items-center gap-2 mb-1"
          >
            <span>{permission.user.username}</span>
            <div className="flex gap-1 items-center">
              <span className="text-sm text-gray-500">
                {levelOptions.find(o => o.value === permission.level)?.label}
              </span>
              <Button
                variant="negative"
                onClick={() =>
                  revokeAccess({
                    variables: { albumId, userId: permission.user.id },
                  })
                }
              >
                {t('general.action.remove', 'Remove')}
              </Button>
            </div>
          </div>
        ))}

        {userOptions.length > 0 && (
          <div className="flex gap-2 items-start mt-2">
            <Dropdown
              className="flex-1"
              items={[
                {
                  value: '',
                  label: t(
                    'sidebar.album.sharing.select_user',
                    'Select a user'
                  ),
                },
                ...userOptions,
              ]}
              selected={newUserId}
              setSelected={setNewUserId}
            />
            <Dropdown
              items={levelOptions}
              selected={newLevel}
              setSelected={value => setNewLevel(value as AlbumPermissionLevel)}
            />
            <Button
              disabled={granting || newUserId === ''}
              onClick={() =>
                grantAccess({
                  variables: { albumId, userId: newUserId, level: newLevel },
                })
              }
            >
              {t('sidebar.album.sharing.share', 'Share')}
            </Button>
          </div>
        )}
        {grantError && (
          <div className="mt-2 text-red-600">{grantError.message}</div>
        )}
        {revokeError && (
          <div className="mt-2 text-red-600">{revokeError.message}</div>
        )}
      </div>
    </SidebarSection>
  )
}

export default SidebarAlbumSharing
