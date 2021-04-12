import React, { useContext } from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { MediaThumbnail, PhotoThumbnail } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import { SidebarContext, UpdateSidebarFn } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'
import { useTranslation } from 'react-i18next'
import { PresentMediaProps_Media } from './presentView/PresentMedia'

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

interface PhotoGalleryProps_Media extends PresentMediaProps_Media {
  thumbnail: null | {
    url: string
    width: number
    height: number
  }
}

type PhotoGalleryProps = {
  loading: boolean
  media: PhotoGalleryProps_Media[]
  activeIndex: number
  presenting: boolean
  onSelectImage(index: number): void
  setPresenting(presenting: boolean): void
  nextImage(): void
  previousImage(): void
  onFavorite(): void
}

const PhotoGallery = ({
  activeIndex = -1,
  media,
  loading,
  onSelectImage,
  presenting,
  setPresenting,
  nextImage,
  previousImage,
  onFavorite,
}: PhotoGalleryProps) => {
  const { t } = useTranslation()
  const { updateSidebar } = useContext(SidebarContext)

  if (
    media === undefined ||
    activeIndex === -1 ||
    media[activeIndex] === undefined
  ) {
    return null
  }

  const activeImage = media[activeIndex]

  const getPhotoElements = (updateSidebar: UpdateSidebarFn) => {
    let photoElements = []
    if (media) {
      photoElements = media.map((photo, index) => {
        const active = activeIndex == index

        return (
          <MediaThumbnail
            key={photo.id}
            media={photo}
            onSelectImage={index => {
              updateSidebar(<MediaSidebar media={photo} />)
              onSelectImage(index)
            }}
            onFavorite={onFavorite}
            setPresenting={setPresenting}
            index={index}
            active={active}
          />
        )
      })
    } else {
      for (let i = 0; i < 6; i++) {
        photoElements.push(<PhotoThumbnail key={i} />)
      }
    }

    return photoElements
  }

  return (
    <ClearWrap>
      <Gallery>
        <Loader active={loading}>
          {t('general.loading.media', 'Loading media')}
        </Loader>
        {getPhotoElements(updateSidebar)}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView
          media={activeImage}
          {...{ nextImage, previousImage, setPresenting }}
        />
      )}
    </ClearWrap>
  )
}

export default PhotoGallery
