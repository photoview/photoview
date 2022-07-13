import React, { useEffect, useReducer } from 'react'
import AlbumTitle from '../album/AlbumTitle'
import MediaGallery, {
  MEDIA_GALLERY_FRAGMENT,
} from '../photoGallery/MediaGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../album/AlbumFilter'
import {
  mediaGalleryReducer,
  urlPresentModeSetupHook,
} from '../photoGallery/mediaGalleryReducer'
import { MediaOrdering, SetOrderingFn } from '../../hooks/useOrderingParams'
import { gql } from '@apollo/client'
import { AlbumGalleryFields } from './__generated__/AlbumGalleryFields'

export const ALBUM_GALLERY_FRAGMENT = gql`
  ${MEDIA_GALLERY_FRAGMENT}

  fragment AlbumGalleryFields on Album {
    id
    title
    subAlbums(order: { order_by: "title", order_direction: $orderDirection }) {
      id
      title
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
    media(
      paginate: { limit: $limit, offset: $offset }
      order: { order_by: $mediaOrderBy, order_direction: $orderDirection }
      onlyFavorites: $onlyFavorites
    ) {
      ...MediaGalleryFields
    }
  }
`

type AlbumGalleryProps = {
  album?: AlbumGalleryFields
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
    const [mediaState, dispatchMedia] = useReducer(mediaGalleryReducer, {
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
