import React, { useEffect, useContext } from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { MediaThumbnail, PhotoThumbnail } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import PropTypes from 'prop-types'
import { SidebarContext } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  min-height: 200px;
  position: relative;
  margin: -4px;
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

const PhotoGallery = ({
  activeIndex = -1,
  media,
  loading,
  onSelectImage,
  presenting,
  setPresenting,
  nextImage,
  previousImage,
}) => {
  const { updateSidebar } = useContext(SidebarContext)

  useEffect(() => {
    const keyDownEvent = e => {
      if (!onSelectImage || activeIndex == -1) {
        return
      }

      if (e.key == 'ArrowRight') {
        nextImage && nextImage()
      }

      if (e.key == 'ArrowLeft') {
        nextImage && previousImage()
      }

      if (e.key == 'Escape' && presenting) {
        setPresenting(false)
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })

  const activeImage = media && activeIndex != -1 && media[activeIndex]

  const getPhotoElements = updateSidebar => {
    let photoElements = []
    if (media) {
      media.filter(media => media.thumbnail)

      photoElements = media.map((photo, index) => {
        const active = activeIndex == index

        let minWidth = 100
        if (photo.thumbnail) {
          minWidth = Math.floor(
            (photo.thumbnail.width / photo.thumbnail.height) * 200
          )
        }

        return (
          <MediaThumbnail
            key={photo.id}
            media={photo}
            onSelectImage={index => {
              updateSidebar(<MediaSidebar media={photo} />)
              onSelectImage(index)
            }}
            setPresenting={setPresenting}
            minWidth={minWidth}
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
    <div>
      <Gallery>
        <Loader active={loading}>Loading images</Loader>
        {getPhotoElements(updateSidebar)}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView
          media={activeImage}
          {...{ nextImage, previousImage, setPresenting }}
        />
      )}
    </div>
  )
}

PhotoGallery.propTypes = {
  loading: PropTypes.bool,
  media: PropTypes.array,
  activeIndex: PropTypes.number,
  presenting: PropTypes.bool,
  onSelectImage: PropTypes.func,
  setPresenting: PropTypes.func,
  nextImage: PropTypes.func,
  previousImage: PropTypes.func,
}

export default PhotoGallery
