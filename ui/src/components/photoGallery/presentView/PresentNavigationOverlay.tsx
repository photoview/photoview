import React, { useState, useRef, useEffect } from 'react'
import styled from 'styled-components'
import { debounce, DebouncedFn } from '../../../helpers/utils'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'

import { useSwipeable } from 'react-swipeable'

import ExitIcon from './icons/Exit'
import InfoIcon from './icons/Info'
import NextIcon from './icons/Next'
import PrevIcon from './icons/Previous'

const StyledOverlayContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const OverlayButton = styled.button`
  width: 64px;
  height: 64px;
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;

  & svg {
    width: 32px;
    height: 32px;
    overflow: visible !important;
  }

  & svg path {
    stroke: rgba(255, 255, 255, 0.5);
    transition-property: stroke, filter;
    transition-duration: 140ms;
  }

  &:hover svg path {
    stroke: rgba(255, 255, 255, 1);
    filter: drop-shadow(0px 0px 2px rgba(0, 0, 0, 0.6));
  }

  /* outline: none above hides the ring for pointer users; keyboard focus
     still needs to be visible, all the more so now that a key press reveals
     the controls and the first stop is a button on top of a photo. The thin
     dark edge keeps the white ring legible over a bright one. */
  &:focus-visible {
    outline: 3px solid rgba(255, 255, 255, 0.95);
    outline-offset: -6px;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.7);
  }

  &:focus-visible svg path {
    stroke: rgba(255, 255, 255, 1);
    filter: drop-shadow(0px 0px 2px rgba(0, 0, 0, 0.6));
  }

  &.hide svg path {
    stroke: rgba(255, 255, 255, 0);
    transition: stroke 300ms;
  }

  /* An invisible button must not stay usable: on touch devices the first tap
     is meant to reveal the controls via the container's click handler, not to
     fire whatever button happens to sit under the finger - and a keyboard
     user must not be able to tab to a control they cannot see, which
     pointer-events alone would still allow. visibility also takes them out
     of the accessibility tree. The delay lets the stroke finish fading out
     first; revealing them again is instant, since no transition applies in
     that direction. */
  &.hide {
    pointer-events: none;
    visibility: hidden;
    transition: visibility 0s 300ms;
  }
`

const ExitButton = styled(OverlayButton)`
  left: 28px;
  top: 28px;
`

const InfoButton = styled(OverlayButton)`
  right: 28px;
  top: 28px;
`

const NavigationButton = styled(OverlayButton)<{ align: 'left' | 'right' }>`
  height: 80%;
  width: 20%;
  top: 10%;

  ${({ align: float }) => (float == 'left' ? 'left: 0;' : null)}
  ${({ align: float }) => (float == 'right' ? 'right: 0;' : null)}

  & svg {
    margin: auto;
    width: 48px;
    height: 64px;
  }
`

type PresentNavigationOverlayProps = {
  children?: React.ReactChild
  dispatchMedia: React.Dispatch<GalleryAction>
  disableSaveCloseInHistory?: boolean
  onInfoClick?: () => void
}

const PresentNavigationOverlay = ({
  children,
  dispatchMedia,
  disableSaveCloseInHistory,
  onInfoClick,
}: PresentNavigationOverlayProps) => {
  const [hide, setHide] = useState(true)
  const onMouseMove = useRef<null | DebouncedFn<() => void>>(null)

  useEffect(() => {
    onMouseMove.current = debounce(
      () => {
        setHide(hide => !hide)
      },
      2000,
      true
    )

    return () => {
      onMouseMove.current?.cancel()
    }
  }, [])

  useEffect(() => {
    // Hidden controls are taken out of the tab order entirely, so a keyboard
    // user has no way to reach them - and no pointer to move. Any key press
    // reveals them, the same as a mouse movement or a tap, which then puts
    // them back within reach of Tab.
    const revealOnKey = () => {
      onMouseMove.current && onMouseMove.current()
    }

    document.addEventListener('keydown', revealOnKey)

    return () => {
      document.removeEventListener('keydown', revealOnKey)
    }
  }, [])

  const handlers = useSwipeable({
    onSwipedLeft: () => dispatchMedia({ type: 'nextImage' }),
    onSwipedRight: () => dispatchMedia({ type: 'previousImage' }),
    preventScrollOnSwipe: true,
    trackMouse: false,
  })

  return (
    <StyledOverlayContainer
      data-testid="present-overlay"
      onMouseMove={() => {
        onMouseMove.current && onMouseMove.current()
      }}
      onClick={() => {
        // Touch devices never fire mousemove, so the controls would
        // otherwise stay hidden forever. A tap toggles them the same way a
        // mouse movement does, and reuses the same auto-hide timer.
        onMouseMove.current && onMouseMove.current()
      }}
    >
      <div {...handlers}>
        {children}
        <NavigationButton
          aria-label="Previous image"
          className={hide ? 'hide' : undefined}
          align="left"
          onClick={() => dispatchMedia({ type: 'previousImage' })}
        >
          <PrevIcon />
        </NavigationButton>
        <NavigationButton
          aria-label="Next image"
          className={hide ? 'hide' : undefined}
          align="right"
          onClick={() => dispatchMedia({ type: 'nextImage' })}
        >
          <NextIcon />
        </NavigationButton>
        <ExitButton
          aria-label="Exit presentation mode"
          className={hide ? 'hide' : undefined}
          onClick={() => {
            if (disableSaveCloseInHistory === true) {
              dispatchMedia({ type: 'closePresentMode' })
            } else {
              closePresentModeAction({ dispatchMedia })
            }
          }}
        >
          <ExitIcon />
        </ExitButton>
        {onInfoClick && (
          <InfoButton
            aria-label="Show media info"
            className={hide ? 'hide' : undefined}
            onClick={onInfoClick}
          >
            <InfoIcon />
          </InfoButton>
        )}
      </div>
    </StyledOverlayContainer>
  )
}

export default PresentNavigationOverlay
