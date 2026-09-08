import React, { useState } from 'react'
import classNames from 'classnames'
import { Link } from 'react-router-dom'
import { ProtectedImage } from '../photoGallery/ProtectedMedia'
import { albumQuery_album_subAlbums } from '../../Pages/AlbumPage/__generated__/albumQuery'
import { useHideAlbumMutation, toggleAlbumHidden } from './albumHideMutations'

interface AlbumBoxImageProps {
  src?: string
}

const AlbumBoxImage = ({ src, ...props }: AlbumBoxImageProps) => {
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
    </div>
  )
}

type AlbumBoxProps = {
  album?: albumQuery_album_subAlbums & { viewerHidden?: boolean }
  customLink?: string
  refetchQueries?: string[]
}

export const AlbumBox = ({
  album,
  customLink,
  refetchQueries,
  ...props
}: AlbumBoxProps) => {
  const wrapperClasses =
    'inline-block text-center text-gray-900 dark:text-gray-200 mx-3 my-2 xs:h-60 xs:w-[220px]'

  const [hideAlbum] = useHideAlbumMutation(refetchQueries)

  if (album) {
    const hidden = album.viewerHidden === true

    return (
      <div className={classNames(wrapperClasses, 'relative group')} {...props}>
        <Link
          to={customLink || `/album/${album.id}`}
          className={classNames('block', { 'opacity-50': hidden })}
        >
          <AlbumBoxImage src={album.thumbnail?.thumbnail?.url} />
          <p className="whitespace-nowrap overflow-hidden overflow-ellipsis">
            {album.title}
          </p>
        </Link>
        <button
          type="button"
          title={hidden ? 'Unhide album' : 'Hide album'}
          className="absolute top-1 right-4 z-10 bg-black/50 text-white rounded-full w-7 h-7 flex items-center justify-center opacity-0 group-hover:opacity-100 focus:opacity-100"
          onClick={e => {
            e.preventDefault()
            e.stopPropagation()
            toggleAlbumHidden(hideAlbum, album.id, hidden)
          }}
        >
          {hidden ? '\u{1F441}' : '\u{1F6AB}'}
        </button>
      </div>
    )
  }

  return (
    <div className={wrapperClasses} {...props}>
      <AlbumBoxImage />
    </div>
  )
}
