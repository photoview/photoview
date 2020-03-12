import React, { useEffect } from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { Photo, PhotoThumbnail } from './Photo'
import { PresentContainer, PresentPhoto } from './PresentView'
import PropTypes from 'prop-types'
import { SidebarConsumer } from '../sidebar/Sidebar'
import PhotoSidebar from '../sidebar/PhotoSidebar'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  min-height: 200px;
  position: relative;
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

export const presentIndexFromHash = hash => {
  let match = hash.match(/present=(\d+)/)
  return match && parseInt(match[1])
}

const PhotoGallery = ({
  activeIndex = -1,
  photos,
  loading,
  onSelectImage,
  presenting,
  setPresenting,
  nextImage,
  previousImage,
}) => {
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

      if (e.key == 'Escape') {
        setPresenting(false)
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })

  const activeImage = photos && activeIndex != -1 && photos[activeIndex]

  const getPhotoElements = updateSidebar => {
    let photoElements = []
    if (photos) {
      photos.filter(photo => photo.thumbnail)

      photoElements = photos.map((photo, index) => {
        const active = activeIndex == index

        let minWidth = 100
        if (photo.thumbnail) {
          minWidth = Math.floor(
            (photo.thumbnail.width / photo.thumbnail.height) * 200
          )
        }

        return (
          <Photo
            key={photo.id}
            photo={photo}
            onSelectImage={index => {
              updateSidebar(<PhotoSidebar photo={photo} />)
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
    <SidebarConsumer>
      {({ updateSidebar }) => (
        <div>
          <Gallery>
            <Loader active={loading}>Loading images</Loader>
            {getPhotoElements(updateSidebar)}
            <PhotoFiller />
          </Gallery>
          {presenting && (
            <PresentContainer>
              <PresentPhoto photo={activeImage} />
            </PresentContainer>
          )}
        </div>
      )}
    </SidebarConsumer>
  )
}

PhotoGallery.propTypes = {
  loading: PropTypes.bool,
  photos: PropTypes.array,
  activeIndex: PropTypes.number,
  presenting: PropTypes.bool,
  onSelectImage: PropTypes.func,
  setPresenting: PropTypes.func,
  nextImage: PropTypes.func,
  previousImage: PropTypes.func,
}

export default PhotoGallery
