import React, { useEffect, useReducer } from 'react'
import AlbumTitle from '../album/AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../album/AlbumFilter'
import { albumQuery_album } from '../../Pages/AlbumPage/__generated__/albumQuery'
import {
  photoGalleryReducer,
  urlPresentModeSetupHook,
} from '../photoGallery/photoGalleryReducer'
import { MediaOrdering, SetOrderingFn } from '../../hooks/useOrderingParams'

type AlbumGalleryProps = {
  album?: albumQuery_album
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
        <PhotoGallery
          loading={loading}
          mediaState={mediaState}
          dispatchMedia={dispatchMedia}
        />
      </div>
    )
  }
)

export default AlbumGallery
