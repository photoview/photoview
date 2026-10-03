import React from 'react'
import { useQuery, gql } from '@apollo/client'
import { SidebarAlbumShare } from './Sharing'
import { useTranslation } from 'react-i18next'
import SidebarHeader from './SidebarHeader'
import {
  getAlbumSidebar,
  getAlbumSidebarVariables,
} from './__generated__/getAlbumSidebar'
import { SidebarAlbumCover } from './AlbumCovers'
import SidebarAlbumDownload from './SidebarDownloadAlbum'
import SidebarAlbumScan from './SidebarAlbumScan'

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
  const { loading, error, data } = useQuery<
    getAlbumSidebar,
    getAlbumSidebarVariables
  >(albumQuery, {
    variables: { id: albumId },
  })

  if (loading) return <div>{t('general.loading.default', 'Loading...')}</div>
  if (error) return <div>{error.message}</div>

  return (
    <div>
      {/* <p>{t('sidebar.album.title', 'Album options')}</p> */}
      <SidebarHeader
        title={
          data?.album.title ??
          t('sidebar.album.title_placeholder', 'Album title')
        }
      />
      <div className="mt-8">
        {/* <h1 className="text-3xl font-semibold">{data.album.title}</h1> */}
        <SidebarAlbumShare id={albumId} />
      </div>
      <div className="mt-8">
        <SidebarAlbumCover id={albumId} />
      </div>
      <div className="mt-8">
        <SidebarAlbumDownload albumID={albumId} />
      </div>
      <div className="mt-8">
        {/* Remounts the mutation state on album change, so a scan still in
            flight for the previous album can't leave this one's button
            stuck disabled. */}
        <SidebarAlbumScan key={albumId} id={albumId} />
      </div>
    </div>
  )
}

export default AlbumSidebar
