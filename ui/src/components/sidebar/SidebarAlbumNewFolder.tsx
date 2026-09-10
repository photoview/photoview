import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import { TextField, Button } from '../../primitives/form/Input'
import {
  createAlbumFolder,
  createAlbumFolderVariables,
} from './__generated__/createAlbumFolder'

export const CREATE_ALBUM_FOLDER_MUTATION = gql`
  mutation createAlbumFolder($parentAlbumId: ID!, $name: String!) {
    createAlbumFolder(parentAlbumId: $parentAlbumId, name: $name) {
      id
      title
    }
  }
`

type SidebarAlbumNewFolderProps = {
  albumId: string
}

const SidebarAlbumNewFolder = ({ albumId }: SidebarAlbumNewFolderProps) => {
  const { t } = useTranslation()
  const [name, setName] = useState('')

  const [createAlbumFolder, { loading, error }] = useMutation<
    createAlbumFolder,
    createAlbumFolderVariables
  >(CREATE_ALBUM_FOLDER_MUTATION, {
    refetchQueries: ['albumQuery'],
    onCompleted: () => setName(''),
    // Without this, a rejected mutation is an unhandled promise rejection -
    // `error` above already renders it.
    onError: () => undefined,
  })

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.new_folder.title', 'New folder')}
      </SidebarSectionTitle>
      <div className="mx-4 flex gap-2 items-start">
        <TextField
          className="flex-1"
          placeholder={t('sidebar.album.new_folder.placeholder', 'Folder name')}
          value={name}
          disabled={loading}
          onChange={e => setName(e.target.value)}
          onKeyUp={e => {
            if (e.key === 'Enter' && name.trim() !== '') {
              createAlbumFolder({
                variables: { parentAlbumId: albumId, name: name.trim() },
              })
            }
          }}
        />
        <Button
          disabled={loading || name.trim() === ''}
          onClick={() =>
            createAlbumFolder({
              variables: { parentAlbumId: albumId, name: name.trim() },
            })
          }
        >
          {t('sidebar.album.new_folder.create', 'Create')}
        </Button>
      </div>
      {error && <div className="mx-4 mt-2 text-red-600">{error.message}</div>}
    </SidebarSection>
  )
}

export default SidebarAlbumNewFolder
