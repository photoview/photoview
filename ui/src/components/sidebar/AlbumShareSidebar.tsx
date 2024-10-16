import React from 'react'
import SidebarAlbumDownload from './SidebarDownloadAlbum'

type AlbumShareSidebarProps = {
  albumId: string
  shareToken?: string
}

const AlbumShareSidebar = ({ albumId, shareToken }: AlbumShareSidebarProps) => {
  return (
    <div>
      <div className="mt-8">
        <SidebarAlbumDownload albumID={albumId} shareToken={shareToken} />
      </div>
    </div>
  )
}

export default AlbumShareSidebar
