import React, { useContext, useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { Button, TextField } from '../../../primitives/form/Input'
import Modal from '../../../primitives/Modal'
import { SidebarContext } from '../Sidebar'
import { renameMedia, renameMediaVariables } from './__generated__/renameMedia'
import { deleteMedia, deleteMediaVariables } from './__generated__/deleteMedia'

export const RENAME_MEDIA_MUTATION = gql`
  mutation renameMedia($mediaId: ID!, $newName: String!) {
    renameMedia(mediaId: $mediaId, newName: $newName) {
      id
      title
      path
    }
  }
`

export const DELETE_MEDIA_MUTATION = gql`
  mutation deleteMedia($mediaId: ID!) {
    deleteMedia(mediaId: $mediaId)
  }
`

type SidebarMediaManageProps = {
  mediaId: string
  currentFileName: string
  albumId: string
}

const SidebarMediaManage = ({
  mediaId,
  currentFileName,
  albumId,
}: SidebarMediaManageProps) => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { updateSidebar } = useContext(SidebarContext)

  const [newName, setNewName] = useState(currentFileName)
  const [showConfirmDelete, setShowConfirmDelete] = useState(false)

  const [renameMedia, { loading: renaming, error: renameError }] = useMutation<
    renameMedia,
    renameMediaVariables
  >(RENAME_MEDIA_MUTATION)

  const [deleteMedia, { loading: deleting, error: deleteError }] = useMutation<
    deleteMedia,
    deleteMediaVariables
  >(DELETE_MEDIA_MUTATION, {
    // No refetchQueries: we're navigating away from this media's own
    // page below, matching SidebarAlbumManage's delete behaviour.
    onCompleted: () => {
      updateSidebar(null)
      navigate(`/album/${albumId}`)
    },
  })

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.media.manage.title', 'Manage file')}
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
              renaming || newName.trim() === '' || newName === currentFileName
            }
            onClick={() =>
              renameMedia({ variables: { mediaId, newName: newName.trim() } })
            }
          >
            {t('sidebar.media.manage.rename', 'Rename')}
          </Button>
        </div>
        {renameError && (
          <div className="mt-2 text-red-600">{renameError.message}</div>
        )}

        <Button
          className="mt-4"
          variant="negative"
          disabled={deleting}
          onClick={() => setShowConfirmDelete(true)}
        >
          {t('sidebar.media.manage.delete', 'Delete this file')}
        </Button>
        {deleteError && (
          <div className="mt-2 text-red-600">{deleteError.message}</div>
        )}
      </div>

      <Modal
        open={showConfirmDelete}
        onClose={() => setShowConfirmDelete(false)}
        title={t('sidebar.media.manage.confirm_delete.title', 'Delete file')}
        description={t(
          'sidebar.media.manage.confirm_delete.description',
          'Move "{{name}}" to the trash? This can be recovered by an administrator, but will disappear from the library immediately.',
          { name: currentFileName }
        )}
        actions={[
          {
            key: 'cancel',
            label: t('general.action.cancel', 'Cancel'),
            onClick: () => setShowConfirmDelete(false),
          },
          {
            key: 'delete',
            label: t('sidebar.media.manage.delete', 'Delete this file'),
            variant: 'negative',
            onClick: () => {
              setShowConfirmDelete(false)
              deleteMedia({ variables: { mediaId } })
            },
          },
        ]}
      />
    </SidebarSection>
  )
}

export default SidebarMediaManage
