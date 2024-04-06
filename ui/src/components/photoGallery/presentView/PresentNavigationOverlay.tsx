import React, { useState, useRef, useEffect } from 'react'
import styled from 'styled-components'
import { debounce, DebouncedFn } from '../../../helpers/utils'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'

import { useSwipeable } from 'react-swipeable'

import ExitIcon from './icons/Exit'
import PlayIcon from './icons/Play'
import PauseIcon from './icons/Pause'
import NextIcon from './icons/Next'
import PrevIcon from './icons/Previous'

const StyledOverlayContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`


const OverlayIconContainer = styled.button`
  width: 64px;
  height: 64px;
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;
  align-items: center;
  justify-content: center;
  display: inline-grid;

  & h1 {
    font-size: 32px;
    font-weight: bolder;
    opacity: 40%;
    display: flex;
    overflow: visible !important;
    transition-property: stroke, filter;
    transition-duration: 140ms;
  }

  &:hover h1 {
    visibility: unset;
    filter: drop-shadow(0px 0px 2px rgba(0, 0, 0, 0.6));
  }

  &.hide h1 {
    visibility: hidden;
    transition: stroke 300ms;
  }
`

const OverlayButton = styled.button`
  width: 64px;
  height: 64px;
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;
  align-items: center;
  justify-content: center;
  display: inline-grid;

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

  &.hide svg path {
    stroke: rgba(255, 255, 255, 0);
    transition: stroke 300ms;
  }
`

const ExitButton = styled(OverlayButton)`
  left: 28px;
  top: 28px;
`
const SlideButton = styled(OverlayButton)<{active:boolean}>`
  left: 92px;
  top: 28px;
`
const IntervalButton = styled(OverlayIconContainer)<{time:integer}>`
  left: 156px;
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
}

const PresentNavigationOverlay = ({
  children,
  dispatchMedia,
  disableSaveCloseInHistory,
}: PresentNavigationOverlayProps) => {
  const [hide, setHide] = useState(true)
  const [slide, setSlide] = useState<boolean>(false)
  const [slideInterval, setSlideInterval] = useState<integer>(3)
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
    const interval = setInterval(() => {
      if (slide) { 
        dispatchMedia({ type: 'nextImage'})
      }
    }, slideInterval*1000)
    return () => {
      clearInterval(interval)
    }
  }, [slide,slideInterval])
  
  const toggle = () => {
    setSlide( (s) => !s ) 
  }
  const toggleSlideInterval = () => {
    setSlideInterval( (s) => (s+1) % 10 == 0 ? 1 : (s+1) % 10  ) 
  }

  const handlers = useSwipeable({
    onSwipedLeft: () => dispatchMedia({ type: 'nextImage' }),
    onSwipedRight: () => dispatchMedia({ type: 'previousImage' }),
    preventScrollOnSwipe: false,
    trackMouse: false,
  })

  return (
    <StyledOverlayContainer
      data-testid="present-overlay"
      onMouseMove={() => {
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
      <SlideButton
        aria-label="Slideshow Button"
        className={hide ? 'hide' : undefined}
        active={slide}
        onClick={toggle}
      >
        {slide ? <PauseIcon /> : <PlayIcon />}
      </SlideButton>
      <IntervalButton
        aria-label="Slideshow Button"
        className={hide ? 'hide' : undefined}
	time={slideInterval}
        onClick={toggleSlideInterval}
      >
        <h1
          className={hide ? 'hide' : undefined}
	> {slideInterval}s </h1>
      </IntervalButton>
    </div>
    </StyledOverlayContainer>
  )
}

export default PresentNavigationOverlay
