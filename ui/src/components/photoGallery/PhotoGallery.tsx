import React, { useContext } from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { MediaThumbnail, PhotoThumbnail } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import { SidebarContext } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'
import { useTranslation } from 'react-i18next'
import { PresentMediaProps_Media } from './presentView/PresentMedia'
import { sidebarPhoto_media_thumbnail } from '../sidebar/__generated__/sidebarPhoto'

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
  thumbnail: sidebarPhoto_media_thumbnail | null
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
  onFavorite?(): void
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

  const activeImage: PhotoGalleryProps_Media | undefined = media[activeIndex]

  let photoElements = []
  if (media) {
    photoElements = media.map((media, index) => {
      const active = activeIndex == index

      return (
        <MediaThumbnail
          key={media.id}
          media={media}
          onSelectImage={index => {
            updateSidebar(<MediaSidebar media={media} />)
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

  return (
    <ClearWrap>
      <Gallery>
        <Loader active={loading}>
          {t('general.loading.media', 'Loading media')}
        </Loader>
        {photoElements}
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
