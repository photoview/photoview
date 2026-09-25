import React from 'react'
import { SidebarAlbumShare } from './Sharing'
import { useTranslation } from 'react-i18next'
import SidebarHeader from './SidebarHeader'
import { authToken } from '../../helpers/authentication'
import { SidebarAlbumCover } from './AlbumCovers'
import SidebarAlbumDownload from './SidebarDownloadAlbum'

type AlbumSidebarProps = {
  albumId: string
  albumTitle: string
}

const AlbumSidebar = ({ albumId, albumTitle }: AlbumSidebarProps) => {
  const { t } = useTranslation()

  return (
    <div>
      {/* <p>{t('sidebar.album.title', 'Album options')}</p> */}
      <SidebarHeader
        title={
          albumTitle ?? t('sidebar.album.title_placeholder', 'Album title')
        }
      />
      {/* don't show albumShare when authToken not available */}
      {authToken() && (
        <div className="mt-8">
          {/* <h1 className="text-3xl font-semibold">{data.album.title}</h1> */}
          <SidebarAlbumShare id={albumId} />
        </div>
      )}
      {/* don't show albumCover when authToken not available */}
      {authToken() && (
        <div className="mt-8">
          <SidebarAlbumCover id={albumId} />
        </div>
      )}
      <div className="mt-8">
        <SidebarAlbumDownload albumID={albumId} />
      </div>
    </div>
  )
}

export default AlbumSidebar
