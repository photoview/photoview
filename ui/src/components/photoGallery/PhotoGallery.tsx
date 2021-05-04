import React from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { MediaThumbnail, PhotoThumbnail } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import { useTranslation } from 'react-i18next'
import { PresentMediaProps_Media } from './presentView/PresentMedia'
import { sidebarPhoto_media_thumbnail } from '../sidebar/__generated__/sidebarPhoto'
import {
  openPresentModeAction,
  PhotoGalleryAction,
  PhotoGalleryState,
} from './photoGalleryReducer'
import {
  toggleFavoriteAction,
  useMarkFavoriteMutation,
} from './photoGalleryMutations'

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

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

const ClearWrap = styled.div`
  clear: both;
`

export interface PhotoGalleryProps_Media extends PresentMediaProps_Media {
  thumbnail: sidebarPhoto_media_thumbnail | null
  favorite?: boolean
}

type PhotoGalleryProps = {
  loading: boolean
  mediaState: PhotoGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
}

const PhotoGallery = ({
  mediaState,
  loading,
  dispatchMedia,
}: PhotoGalleryProps) => {
  const { t } = useTranslation()

  const [markFavorite] = useMarkFavoriteMutation()

  const { media, activeIndex, presenting } = mediaState

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
      photoElements.push(<PhotoThumbnail key={i} />)
    }
  }

  return (
    <ClearWrap>
      <Gallery data-testid="photo-gallery-wrapper">
        <Loader active={loading}>
          {t('general.loading.media', 'Loading media')}
        </Loader>
        {photoElements}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView
          activeMedia={mediaState.media[mediaState.activeIndex]}
          dispatchMedia={dispatchMedia}
        />
      )}
    </ClearWrap>
  )
}

export default PhotoGallery
