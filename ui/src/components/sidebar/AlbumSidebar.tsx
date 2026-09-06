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
import SidebarAlbumNewFolder from './SidebarAlbumNewFolder'
import SidebarAlbumUpload from './SidebarAlbumUpload'
import SidebarAlbumManage from './SidebarAlbumManage'
import SidebarAlbumSharing from './SidebarAlbumSharing'

const albumQuery = gql`
  query getAlbumSidebar($id: ID!) {
    album(id: $id) {
      id
      title
      viewerCanUpload
      viewerCanDelete
      viewerIsOwner
      parentAlbumId
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
        {/* Remounts the mutation state on album change, so a scan promise
            still in flight for the previous album can't leave this
            album's button stuck disabled. */}
        <SidebarAlbumScan key={albumId} id={albumId} />
      </div>
      {data?.album.viewerCanUpload && (
        <>
          <div className="mt-8">
            <SidebarAlbumNewFolder key={albumId} albumId={albumId} />
          </div>
          <div className="mt-8">
            <SidebarAlbumUpload key={albumId} albumId={albumId} />
          </div>
        </>
      )}
      {data?.album.parentAlbumId && data.album.viewerCanDelete && (
        <div className="mt-8">
          <SidebarAlbumManage
            key={albumId}
            albumId={albumId}
            albumTitle={data.album.title}
          />
        </div>
      )}
      {data?.album.viewerIsOwner && (
        <div className="mt-8">
          <SidebarAlbumSharing key={albumId} albumId={albumId} />
        </div>
      )}
    </div>
  )
}

export default AlbumSidebar
