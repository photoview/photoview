import React, { useEffect, useReducer } from 'react'
import AlbumTitle from '../album/AlbumTitle'
import MediaGallery from '../photoGallery/MediaGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../album/AlbumFilter'
import {
  albumQuery_album_media_highRes,
  albumQuery_album_media_thumbnail,
  albumQuery_album_media_videoWeb,
  albumQuery_album_subAlbums,
} from '../../Pages/AlbumPage/__generated__/albumQuery'
import {
  photoGalleryReducer,
  urlPresentModeSetupHook,
} from '../photoGallery/photoGalleryReducer'
import { MediaOrdering, SetOrderingFn } from '../../hooks/useOrderingParams'
import { MediaType } from '../../__generated__/globalTypes'

type AlbumGalleryAlbum = {
  __typename: 'Album'
  id: string
  title: string
  subAlbums: albumQuery_album_subAlbums[]
  media: {
    __typename: 'Media'
    id: string
    type: MediaType
    /**
     * URL to display the media in a smaller resolution
     */
    thumbnail: albumQuery_album_media_thumbnail | null
    /**
     * URL to display the photo in full resolution, will be null for videos
     */
    highRes: albumQuery_album_media_highRes | null
    /**
     * URL to get the video in a web format that can be played in the browser, will be null for photos
     */
    videoWeb: albumQuery_album_media_videoWeb | null
    favorite?: boolean
    blurhash: string | null
  }[]
}

type AlbumGalleryProps = {
  album?: AlbumGalleryAlbum
  loading?: boolean
  customAlbumLink?(albumID: string): string
  showFilter?: boolean
  setOnlyFavorites?(favorites: boolean): void
  setOrdering?: SetOrderingFn
  ordering?: MediaOrdering
  onlyFavorites?: boolean
  onFavorite?(): void
}

const AlbumGallery = React.forwardRef(
  (
    {
      album,
      loading = false,
      customAlbumLink,
      showFilter = false,
      setOnlyFavorites,
      setOrdering,
      ordering,
      onlyFavorites = false,
    }: AlbumGalleryProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    const [mediaState, dispatchMedia] = useReducer(photoGalleryReducer, {
      presenting: false,
      activeIndex: -1,
      media: album?.media || [],
    })

    useEffect(() => {
      dispatchMedia({ type: 'replaceMedia', media: album?.media || [] })
    }, [album?.media])

    urlPresentModeSetupHook({
      dispatchMedia,
      openPresentMode: event => {
        dispatchMedia({
          type: 'openPresentMode',
          activeIndex: event.state.activeIndex,
        })
      },
    })

    let subAlbumElement = null
    if (album) {
      if (album.subAlbums.length > 0) {
        subAlbumElement = (
          <AlbumBoxes
            albums={album.subAlbums}
            getCustomLink={customAlbumLink}
          />
        )
      }
    } else {
      subAlbumElement = <AlbumBoxes />
    }

    return (
      <div ref={ref}>
        {showFilter && (
          <AlbumFilter
            onlyFavorites={onlyFavorites}
            setOnlyFavorites={setOnlyFavorites}
            setOrdering={setOrdering}
            ordering={ordering}
          />
        )}
        <AlbumTitle album={album} disableLink />
        {subAlbumElement}
        <MediaGallery
          loading={loading}
          mediaState={mediaState}
          dispatchMedia={dispatchMedia}
        />
      </div>
    )
  }
)

export default AlbumGallery
