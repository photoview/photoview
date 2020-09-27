import React, { useEffect } from 'react'
import PropTypes from 'prop-types'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentMedia from './PresentMedia'

const StyledContainer = styled.div`
  position: fixed;
  width: 100vw;
  height: 100vh;
  background-color: black;
  color: white;
  top: 0;
  left: 0;
  z-index: 100;
`

const PreventScroll = createGlobalStyle`
  * {
    overflow: hidden !important;
  }
`

const PresentView = ({
  className,
  media,
  imageLoaded,
  nextImage,
  previousImage,
  setPresenting,
}) => {
  useEffect(() => {
    const keyDownEvent = e => {
      if (e.key == 'ArrowRight') {
        nextImage && nextImage()
        e.stopPropagation()
      }

      if (e.key == 'ArrowLeft') {
        nextImage && previousImage()
        e.stopPropagation()
      }

      if (e.key == 'Escape') {
        setPresenting(false)
        e.stopPropagation()
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })

  return (
    <StyledContainer {...className}>
      <PreventScroll />
      <PresentNavigationOverlay
        {...{ nextImage, previousImage, setPresenting }}
      >
        <PresentMedia media={media} imageLoaded={imageLoaded} />
      </PresentNavigationOverlay>
    </StyledContainer>
  )
}

PresentView.propTypes = {
  media: PropTypes.object.isRequired,
  imageLoaded: PropTypes.func,
  className: PropTypes.string,
  nextImage: PropTypes.func.isRequired,
  previousImage: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
}

export default PresentView
