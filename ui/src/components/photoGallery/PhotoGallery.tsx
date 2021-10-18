import React, { useContext } from 'react'
import styled from 'styled-components'
import { MediaThumbnail, MediaPlaceholder } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import { PresentMediaProps_Media } from './presentView/PresentMedia'
import {
  openPresentModeAction,
  PhotoGalleryAction,
  PhotoGalleryState,
} from './photoGalleryReducer'
import {
  toggleFavoriteAction,
  useMarkFavoriteMutation,
} from './photoGalleryMutations'
import MediaSidebar from '../sidebar/MediaSidebar/MediaSidebar'
import { SidebarContext } from '../sidebar/Sidebar'
import { sidebarMediaQuery_media_thumbnail } from '../sidebar/MediaSidebar/__generated__/sidebarMediaQuery'

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

export interface PhotoGalleryProps_Media extends PresentMediaProps_Media {
  thumbnail: sidebarMediaQuery_media_thumbnail | null
  favorite?: boolean
}

type PhotoGalleryProps = {
  loading: boolean
  mediaState: PhotoGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
}

const PhotoGallery = ({ mediaState, dispatchMedia }: PhotoGalleryProps) => {
  const [markFavorite] = useMarkFavoriteMutation()

  const { media, activeIndex, presenting } = mediaState

  const { updateSidebar } = useContext(SidebarContext)

  let photoElements = []
  if (media) {
    photoElements = media.map((media, index) => {
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
            openPresentModeAction({ dispatchMedia, activeIndex: index })
          }}
        />
      )
    })
  } else {
    for (let i = 0; i < 6; i++) {
      photoElements.push(<MediaPlaceholder key={i} />)
    }
  }

  return (
    <>
      <Gallery data-testid="photo-gallery-wrapper">
        {photoElements}
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

export default PhotoGallery
