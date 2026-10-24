import React, { useContext, useEffect, useRef, useState } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentMedia from './PresentMedia'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'
import { SidebarContext } from '../../sidebar/Sidebar'
import MediaSidebar from '../../sidebar/MediaSidebar/MediaSidebar'

const StyledContainer = styled.div<{ $besidePanel: boolean }>`
  position: fixed;
  width: 100vw;
  height: 100vh;
  background-color: black;
  color: white;
  top: 0;
  left: 0;
  z-index: 100;
  overscroll-behavior: none;

  /* With the info panel pinned, a wide screen gives the photo and its
     controls the space beside the panel instead of hiding part of both under
     it - the same as a pinned sidebar on the album and timeline pages. The
     420px is the sidebar's width at this breakpoint (lg:w-[420px]). */
  @media (min-width: 1024px) {
    ${({ $besidePanel }) => ($besidePanel ? 'width: calc(100vw - 420px);' : '')}
  }
`

// Locks scrolling on the page behind the fullscreen viewer. Scoped to
// html/body rather than every element, so panels rendered on top of the
// viewer (e.g. the media info sidebar) can still scroll their own content.
const PreventScroll = createGlobalStyle`
  html, body {
    overflow: hidden !important;
  }

  /* The media info panel is the shared sidebar, which normally sits at z-40 -
     below this fullscreen view at z-100, so opening it from here changed its
     state without ever showing it. Lifting it only while the viewer is mounted
     leaves its stacking everywhere else exactly as it was. */
  [data-sidebar] {
    z-index: 110 !important;
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
  const {
    updateSidebar,
    setPinned,
    pinned,
    content: sidebarContent,
  } = useContext(SidebarContext)
  const [infoOpen, setInfoOpen] = useState(false)

  // Read by the unmount cleanup below, which would otherwise only ever see the
  // value infoOpen had on the first render.
  const infoOpenRef = useRef(false)
  useEffect(() => {
    infoOpenRef.current = infoOpen
  }, [infoOpen])

  useEffect(
    () => () => {
      // Leaving the viewer with the info panel open would otherwise leave it
      // behind over the gallery, still describing the last presented photo. A
      // sidebar the user opened before presenting is theirs and stays open.
      if (infoOpenRef.current) {
        updateSidebar(null)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

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
    <StyledContainer
      className={className}
      $besidePanel={infoOpen && pinned && sidebarContent != null}
    >
      <PreventScroll />
      <PresentNavigationOverlay
        dispatchMedia={dispatchMedia}
        disableSaveCloseInHistory={disableSaveCloseInHistory}
        onInfoClick={() => {
          setInfoOpen(true)
          updateSidebar(<MediaSidebar media={activeMedia} />)
          // Pinned, so a wide screen lays the panel out beside the photo.
          // Unpinning it drops back to the panel overlaying the photo, and on
          // a phone, where pinning has no layout of its own, it overlays
          // either way.
          setPinned(true)
        }}
      >
        <PresentMedia media={activeMedia} imageLoaded={imageLoaded} />
      </PresentNavigationOverlay>
    </StyledContainer>
  )
}

export default PresentView
