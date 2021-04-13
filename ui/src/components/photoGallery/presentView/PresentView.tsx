import React, { useEffect } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentMedia, { PresentMediaProps_Media } from './PresentMedia'

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

type PresentViewProps = {
  media: PresentMediaProps_Media
  className?: string
  imageLoaded?(): void
  nextImage(): void
  previousImage(): void
  setPresenting(presenting: boolean): void
}

const PresentView = ({
  className,
  media,
  imageLoaded,
  nextImage,
  previousImage,
  setPresenting,
}: PresentViewProps) => {
  useEffect(() => {
    const keyDownEvent = (e: KeyboardEvent) => {
      if (e.key == 'ArrowRight') {
        nextImage()
        e.stopPropagation()
      }

      if (e.key == 'ArrowLeft') {
        previousImage()
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

export default PresentView
