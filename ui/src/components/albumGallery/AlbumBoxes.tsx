import React from 'react'
import { albumQuery_album_subAlbums } from '../../Pages/AlbumPage/__generated__/albumQuery'
import { AlbumBox } from './AlbumBox'

type AlbumBoxesProps = {
  error?: Error
  albums?: albumQuery_album_subAlbums[]
  getCustomLink?(albumID: string): string
}

const AlbumBoxes = ({ error, albums, getCustomLink }: AlbumBoxesProps) => {
  if (error) return <div>Error {error.message}</div>

  let albumElements = []

  if (albums !== undefined) {
    albumElements = albums.map(album => (
      <AlbumBox
        key={album.id}
        album={album}
        customLink={getCustomLink ? getCustomLink(album.id) : undefined}
      />
    ))
  } else {
    for (let i = 0; i < 4; i++) {
      albumElements.push(<AlbumBox key={i} />)
    }
  }

  return <div className="-mx-3 my-6">{albumElements}</div>
}

export default AlbumBoxes
