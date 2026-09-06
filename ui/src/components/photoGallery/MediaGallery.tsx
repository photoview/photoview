import React, { useContext } from 'react'
import styled from 'styled-components'
import { MediaThumbnail, MediaPlaceholder } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import {
  openPresentModeAction,
  PhotoGalleryAction,
  MediaGalleryState,
} from './mediaGalleryReducer'
import {
  toggleFavoriteAction,
  useMarkFavoriteMutation,
} from './photoGalleryMutations'
import MediaSidebar from '../sidebar/MediaSidebar/MediaSidebar'
import { SidebarContext } from '../sidebar/Sidebar'
import { gql } from '@apollo/client'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  min-height: 200px;
  position: relative;
  margin: -4px;

  @media (max-width: 1000px) {
    /* Compensate for tab bar on mobile */
    margin-bottom: 76px;
  }
`

export const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

export const MEDIA_GALLERY_FRAGMENT = gql`
  fragment MediaGalleryFields on Media {
    id
    type
    blurhash
    thumbnail {
      url
      width
      height
    }
    highRes {
      url
    }
    videoWeb {
      url
    }
    favorite
  }
`

type MediaGalleryProps = {
  loading: boolean
  mediaState: MediaGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
  // Distinguishes this gallery's presentation history from any other
  // MediaGallery on the same page. See urlPresentModeSetupHook.
  groupId?: string
  selectMode?: boolean
  selectedIds?: Set<string>
  onToggleSelect?(mediaId: string): void
}

const MediaGallery = ({
  mediaState,
  dispatchMedia,
  groupId,
  selectMode,
  selectedIds,
  onToggleSelect,
}: MediaGalleryProps) => {
  const [markFavorite] = useMarkFavoriteMutation()

  const { media, activeIndex, presenting } = mediaState

  const { updateSidebar } = useContext(SidebarContext)

  let mediaElements = []
  if (media) {
    mediaElements = media.map((media, index) => {
      const active = activeIndex == index

      return (
        <MediaThumbnail
          key={media.id}
          media={media}
          active={active}
          selectImage={() => {
            dispatchMedia({
              type: 'selectImage',
              index,
            })
            updateSidebar(<MediaSidebar media={mediaState.media[index]} />)
          }}
          clickFavorite={() => {
            toggleFavoriteAction({
              media,
              markFavorite,
            })
          }}
          clickPresent={() => {
            openPresentModeAction({
              dispatchMedia,
              activeIndex: index,
              groupId,
            })
          }}
          selectMode={selectMode}
          selected={selectedIds?.has(media.id)}
          onToggleSelect={() => onToggleSelect?.(media.id)}
        />
      )
    })
  } else {
    for (let i = 0; i < 6; i++) {
      mediaElements.push(<MediaPlaceholder key={i} />)
    }
  }

  return (
    <>
      <Gallery data-testid="photo-gallery-wrapper">
        {mediaElements}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView
          activeMedia={mediaState.media[mediaState.activeIndex]}
          dispatchMedia={dispatchMedia}
        />
      )}
    </>
  )
}

export default MediaGallery
