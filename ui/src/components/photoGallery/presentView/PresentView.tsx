import React, { useContext, useEffect, useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentMedia from './PresentMedia'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'
import { SidebarContext } from '../../sidebar/Sidebar'
import MediaSidebar from '../../sidebar/MediaSidebar/MediaSidebar'

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

// Locks scrolling on the page behind the fullscreen viewer. Scoped to
// html/body rather than every element, so panels rendered on top of the
// viewer (e.g. the media info sidebar) can still scroll their own content.
const PreventScroll = createGlobalStyle`
  html, body {
    overflow: hidden !important;
  }
`

type PresentViewProps = {
  className?: string
  imageLoaded?(): void
  activeMedia: MediaGalleryFields
  dispatchMedia: React.Dispatch<GalleryAction>
  disableSaveCloseInHistory?: boolean
}

const PresentView = ({
  className,
  imageLoaded,
  activeMedia,
  dispatchMedia,
  disableSaveCloseInHistory,
}: PresentViewProps) => {
  const { updateSidebar, content: sidebarContent } = useContext(SidebarContext)
  const [infoOpen, setInfoOpen] = useState(false)

  useEffect(() => {
    // The sidebar was closed some other way (e.g. its own close button),
    // so navigating to another image shouldn't reopen it.
    if (sidebarContent === null) {
      setInfoOpen(false)
    }
  }, [sidebarContent])

  useEffect(() => {
    // Keep an already-open info panel in sync with the active image - it
    // was built from activeMedia at the moment the panel was opened, and
    // otherwise keeps showing that same image after navigating away.
    if (infoOpen) {
      updateSidebar(<MediaSidebar media={activeMedia} />)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeMedia])

  useEffect(() => {
    const keyDownEvent = (e: KeyboardEvent) => {
      if (e.key == 'ArrowRight') {
        e.stopPropagation()
        dispatchMedia({ type: 'nextImage' })
      }

      if (e.key == 'ArrowLeft') {
        e.stopPropagation()
        dispatchMedia({ type: 'previousImage' })
      }

      if (e.key == 'Escape') {
        e.stopPropagation()

        if (disableSaveCloseInHistory === true) {
          dispatchMedia({ type: 'closePresentMode' })
        } else {
          closePresentModeAction({ dispatchMedia })
        }
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })

  return (
    <StyledContainer className={className}>
      <PreventScroll />
      <PresentNavigationOverlay
        dispatchMedia={dispatchMedia}
        disableSaveCloseInHistory
        onInfoClick={() => {
          setInfoOpen(true)
          updateSidebar(<MediaSidebar media={activeMedia} />)
        }}
      >
        <PresentMedia media={activeMedia} imageLoaded={imageLoaded} />
      </PresentNavigationOverlay>
    </StyledContainer>
  )
}

export default PresentView
