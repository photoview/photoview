import React, { useContext, useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import { Button, TextField } from '../../primitives/form/Input'
import Dropdown from '../../primitives/form/Dropdown'
import Modal from '../../primitives/Modal'
import { SidebarContext } from './Sidebar'
import { sidebarManageAlbumsQuery } from './__generated__/sidebarManageAlbumsQuery'
import { renameAlbum, renameAlbumVariables } from './__generated__/renameAlbum'
import { moveAlbum, moveAlbumVariables } from './__generated__/moveAlbum'
import { deleteAlbum, deleteAlbumVariables } from './__generated__/deleteAlbum'

export const MANAGE_ALBUMS_QUERY = gql`
  query sidebarManageAlbumsQuery {
    myAlbums(onlyRoot: false, showEmpty: true) {
      id
      title
    }
  }
`

export const RENAME_ALBUM_MUTATION = gql`
  mutation renameAlbum($albumId: ID!, $newName: String!) {
    renameAlbum(albumId: $albumId, newName: $newName) {
      id
      title
    }
  }
`

export const MOVE_ALBUM_MUTATION = gql`
  mutation moveAlbum($albumId: ID!, $newParentAlbumId: ID!) {
    moveAlbum(albumId: $albumId, newParentAlbumId: $newParentAlbumId) {
      id
    }
  }
`

export const DELETE_ALBUM_MUTATION = gql`
  mutation deleteAlbum($albumId: ID!) {
    deleteAlbum(albumId: $albumId)
  }
`

type SidebarAlbumManageProps = {
  albumId: string
  albumTitle: string
}

const SidebarAlbumManage = ({
  albumId,
  albumTitle,
}: SidebarAlbumManageProps) => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { updateSidebar } = useContext(SidebarContext)

  const [newName, setNewName] = useState(albumTitle)
  const [destination, setDestination] = useState('')
  const [showConfirmDelete, setShowConfirmDelete] = useState(false)

  const { data } = useQuery<sidebarManageAlbumsQuery>(MANAGE_ALBUMS_QUERY)

  const [renameAlbum, { loading: renaming, error: renameError }] = useMutation<
    renameAlbum,
    renameAlbumVariables
  >(RENAME_ALBUM_MUTATION, {
    refetchQueries: ['albumQuery'],
    // Without this, a rejected mutation is an unhandled promise rejection -
    // renameError above already renders it, this just avoids the console
    // warning.
    onError: () => undefined,
  })

  const [moveAlbum, { loading: moving, error: moveError }] = useMutation<
    moveAlbum,
    moveAlbumVariables
  >(MOVE_ALBUM_MUTATION, {
    refetchQueries: ['albumQuery'],
    onCompleted: () => navigate(`/album/${destination}`),
    // Without this, a rejected mutation is an unhandled promise rejection -
    // moveError above already renders it, this just avoids the console
    // warning.
    onError: () => undefined,
  })

  const [deleteAlbum, { loading: deleting, error: deleteError }] = useMutation<
    deleteAlbum,
    deleteAlbumVariables
  >(DELETE_ALBUM_MUTATION, {
    // No refetchQueries here: we're navigating away from the deleted
    // album's own page below, and refetching its still-mounted albumQuery
    // first would surface a harmless but confusing "album not found" error
    // toast before that navigation completes.
    onCompleted: () => {
      updateSidebar(null)
      navigate('/albums')
    },
    // Without this, a rejected mutation is an unhandled promise
    // rejection - deleteError above already renders it.
    onError: () => undefined,
  })

  const destinationOptions = (data?.myAlbums ?? [])
    .filter(album => album.id !== albumId)
    .map(album => ({ value: album.id, label: album.title }))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.manage.title', 'Manage folder')}
      </SidebarSectionTitle>
      <div className="mx-4">
        <div className="flex gap-2 items-start">
          <TextField
            className="flex-1"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            disabled={renaming}
          />
          <Button
            disabled={
              renaming || newName.trim() === '' || newName === albumTitle
            }
            onClick={() =>
              renameAlbum({
                variables: { albumId, newName: newName.trim() },
              })
            }
          >
            {t('sidebar.album.manage.rename', 'Rename')}
          </Button>
        </div>
        {renameError && (
          <div className="mt-2 text-red-600">{renameError.message}</div>
        )}

        <div className="mt-4">
          <Dropdown
            className="w-full"
            items={destinationOptions}
            selected={destination}
            setSelected={setDestination}
          />
          <Button
            className="mt-2"
            disabled={moving || destination === ''}
            onClick={() =>
              moveAlbum({
                variables: { albumId, newParentAlbumId: destination },
              })
            }
          >
            {t('sidebar.album.manage.move', 'Move')}
          </Button>
        </div>
        {moveError && (
          <div className="mt-2 text-red-600">{moveError.message}</div>
        )}

        <Button
          className="mt-4"
          variant="negative"
          disabled={deleting}
          onClick={() => setShowConfirmDelete(true)}
        >
          {t('sidebar.album.manage.delete', 'Delete this folder')}
        </Button>
        {deleteError && (
          <div className="mt-2 text-red-600">{deleteError.message}</div>
        )}
      </div>

      <Modal
        open={showConfirmDelete}
        onClose={() => setShowConfirmDelete(false)}
        title={t('sidebar.album.manage.confirm_delete.title', 'Delete folder')}
        description={t(
          'sidebar.album.manage.confirm_delete.description',
          'Move "{{title}}" and everything inside it to the trash? This can be recovered by an administrator, but will disappear from the library immediately.',
          { title: albumTitle }
        )}
        actions={[
          {
            key: 'cancel',
            label: t('general.action.cancel', 'Cancel'),
            onClick: () => setShowConfirmDelete(false),
          },
          {
            key: 'delete',
            label: t('sidebar.album.manage.delete', 'Delete this folder'),
            variant: 'negative',
            onClick: () => {
              setShowConfirmDelete(false)
              deleteAlbum({ variables: { albumId } })
            },
          },
        ]}
      />
    </SidebarSection>
  )
}

export default SidebarAlbumManage
