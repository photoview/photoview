import React from 'react'
import { useQuery, gql } from '@apollo/client'
import { SidebarAlbumShare } from './Sharing'
import { useTranslation } from 'react-i18next'

const albumQuery = gql`
  query getAlbumSidebar($id: ID!) {
    album(id: $id) {
      id
      title
    }
  }
`

type AlbumSidebarProps = {
  albumId: string
}

const AlbumSidebar = ({ albumId }: AlbumSidebarProps) => {
  const { t } = useTranslation()
  const { loading, error, data } = useQuery(albumQuery, {
    variables: { id: albumId },
  })

  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>
  if (error) return <div>{error.message}</div>

  return (
    <div>
      <p>{t('sidebar.album.title', 'Album options')}</p>
      <div>
        <h1>{data.album.title}</h1>
        <SidebarAlbumShare id={data.album.id} />
      </div>
    </div>
  )
}

export default AlbumSidebar
