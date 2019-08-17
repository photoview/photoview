import React, { useState, useEffect } from 'react'
import { animated } from 'react-spring'
import { Transition } from 'react-spring/renderprops'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { Photo } from './Photo'
import { PresentContainer, PresentPhoto } from './PresentView'
import PropTypes from 'prop-types'
import { SidebarConsumer } from '../sidebar/Sidebar'
import PhotoSidebar from '../sidebar/PhotoSidebar'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
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
        setMoveDirection('right')
        nextImage && nextImage()
      }

      if (e.key == 'ArrowLeft') {
        setMoveDirection('left')
        nextImage && previousImage()
      }

      if (e.key == 'Escape') {
        setMoveDirection(null)
        setPresenting(false)
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })

  const [moveDirection, setMoveDirection] = useState(null)

  const activeImage = photos && activeIndex != -1 && photos[activeIndex]

  const getPhotoElements = updateSidebar => {
    let photoElements = null
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
    }

    return photoElements
  }

  let transformDirectionIndex = 0
  if (moveDirection == 'right') transformDirectionIndex = 1
  if (moveDirection == 'left') transformDirectionIndex = 2

  const presentViewTransitionConfig = {
    items: activeImage,
    keys: x => x,
    config: {
      tension: 220,
    },
    from: {
      opacity: 0,
      transform: [
        'translate(0%, 0)',
        'translate(12%, 0)',
        'translate(-12%, 0)',
      ][transformDirectionIndex],
    },
    enter: {
      opacity: 1,
      transform: 'translate(0%, 0)',
    },
  }

  const AnimatedPresentPhoto = animated(PresentPhoto)

  return (
    <SidebarConsumer>
      {({ updateSidebar }) => (
        <div>
          {!presenting ? (
            <Gallery>
              <Loader active={loading}>Loading images</Loader>
              {getPhotoElements(updateSidebar)}
              <PhotoFiller />
            </Gallery>
          ) : (
            <PresentContainer>
              <Transition {...presentViewTransitionConfig}>
                {photo => props => (
                  <PresentPhoto
                    thumbnail={photo && photo.thumbnail.url}
                    photo={photo}
                    style={props}
                  />
                )}
              </Transition>
            </PresentContainer>
          )}
        </div>
      )}
    </SidebarConsumer>
  )
  // }
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
