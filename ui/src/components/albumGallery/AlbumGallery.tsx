import React, { useContext, useEffect, useReducer } from 'react'
import AlbumTitle from '../AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../AlbumFilter'
import { albumQuery_album } from '../../Pages/AlbumPage/__generated__/albumQuery'
import { OrderDirection } from '../../../__generated__/globalTypes'
import {
  photoGalleryReducer,
  urlPresentModeSetupHook,
} from '../photoGallery/photoGalleryReducer'
import { SidebarContext } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'

type AlbumGalleryProps = {
  album?: albumQuery_album
  loading?: boolean
  customAlbumLink?(albumID: string): string
  showFilter?: boolean
  setOnlyFavorites?(favorites: boolean): void
  setOrdering?(ordering: { orderBy: string }): void
  ordering?: { orderBy: string | null; orderDirection: OrderDirection | null }
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
    const { updateSidebar } = useContext(SidebarContext)

    const [mediaState, dispatchMedia] = useReducer(photoGalleryReducer, {
      presenting: false,
      activeIndex: -1,
      media: album?.media || [],
    })

    useEffect(() => {
      dispatchMedia({ type: 'replaceMedia', media: album?.media || [] })
    }, [album?.media])

    useEffect(() => {
      if (mediaState.activeIndex != -1) {
        updateSidebar(
          <MediaSidebar media={mediaState.media[mediaState.activeIndex]} />
        )
      } else {
        updateSidebar(null)
      }
    }, [mediaState.activeIndex])

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
            loading={loading}
            albums={album.subAlbums}
            getCustomLink={customAlbumLink}
          />
        )
      }
    } else {
      subAlbumElement = <AlbumBoxes loading={loading} />
    }

    return (
      <div ref={ref}>
        <AlbumTitle album={album} disableLink />
        {showFilter && (
          <AlbumFilter
            onlyFavorites={onlyFavorites}
            setOnlyFavorites={setOnlyFavorites}
            setOrdering={setOrdering}
            ordering={ordering}
          />
        )}
        {subAlbumElement}
        {
          <h2
            style={{
              opacity: loading ? 0 : 1,
              display: album && album.subAlbums.length > 0 ? 'block' : 'none',
            }}
          >
            Images
          </h2>
        }
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
