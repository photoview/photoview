import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { ProtectedImage } from '../photoGallery/ProtectedMedia'
import { albumQuery_album_subAlbums } from '../../Pages/AlbumPage/__generated__/albumQuery'

const NEW_ALBUM_DAYS = 14

function isNewAlbum(createdAt?: string): boolean {
  if (!createdAt) return false
  const created = new Date(createdAt)
  const now = new Date()
  const diffMs = now.getTime() - created.getTime()
  const diffDays = diffMs / (1000 * 60 * 60 * 24)
  return diffDays <= NEW_ALBUM_DAYS
}

interface AlbumBoxImageProps {
  src?: string
  isNew?: boolean
}

const AlbumBoxImage = ({ src, isNew, ...props }: AlbumBoxImageProps) => {
  const [loaded, setLoaded] = useState(false)

  let image = null
  if (src) {
    image = (
      <ProtectedImage
        className="object-cover object-center w-full h-full rounded-lg"
        {...props}
        onLoad={() => setLoaded(true)}
        src={src}
      />
    )
  }

  let placeholder = null
  if (!loaded) {
    placeholder = (
      <div className="bg-gray-100 dark:bg-[#191c1f] animate-pulse w-full h-full rounded-lg absolute top-0"></div>
    )
  }

  return (
    <div className="xs:w-[220px] xs:h-[220px] relative rounded-lg">
      {image}
      {placeholder}
      {isNew && (
        <div className="absolute top-2 right-2 z-10">
          <span className="bg-green-500 text-white text-xs font-bold px-2 py-0.5 rounded shadow-md">
            NEW
          </span>
        </div>
      )}
    </div>
  )
}

type AlbumBoxProps = {
  album?: albumQuery_album_subAlbums
  customLink?: string
}

export const AlbumBox = ({ album, customLink, ...props }: AlbumBoxProps) => {
  const wrapperClasses =
    'inline-block text-center text-gray-900 dark:text-gray-200 mx-3 my-2 xs:h-60 xs:w-[220px]'

  if (album) {
    return (
      <Link
        to={customLink || `/album/${album.id}`}
        className={wrapperClasses}
        {...props}
      >
        <AlbumBoxImage src={album.thumbnail?.thumbnail?.url} isNew={isNewAlbum(album.createdAt)} />
        <p className="whitespace-nowrap overflow-hidden overflow-ellipsis">
          {album.title}
        </p>
      </Link>
    )
  }

  return (
    <div className={wrapperClasses} {...props}>
      <AlbumBoxImage />
    </div>
  )
}
