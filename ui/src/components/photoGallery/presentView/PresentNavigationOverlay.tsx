import React, { useState, useRef, useEffect } from 'react'
import styled from 'styled-components'
import { debounce, DebouncedFn } from '../../../helpers/utils'
import { PhotoGalleryAction } from '../photoGalleryReducer'

import ExitIcon from './icons/Exit'
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

  &.hide svg path {
    stroke: rgba(255, 255, 255, 0);
    transition: stroke 300ms;
  }
`

const ExitButton = styled(OverlayButton)`
  left: 28px;
  top: 28px;
`

const NavigationButton = styled(OverlayButton)<{ float: 'left' | 'right' }>`
  height: 80%;
  width: 20%;
  top: 10%;

  ${({ float }) => (float == 'left' ? 'left: 0;' : null)}
  ${({ float }) => (float == 'right' ? 'right: 0;' : null)}

  & svg {
    width: 48px;
    height: 64px;
  }
`

type PresentNavigationOverlayProps = {
  children?: React.ReactChild
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
}

const PresentNavigationOverlay = ({
  children,
  dispatchMedia,
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

  return (
    <StyledOverlayContainer
      onMouseMove={() => {
        onMouseMove.current && onMouseMove.current()
      }}
    >
      {children}
      <NavigationButton
        className={hide ? 'hide' : undefined}
        float="left"
        onClick={() => dispatchMedia({ type: 'previousImage' })}
      >
        <PrevIcon />
      </NavigationButton>
      <NavigationButton
        className={hide ? 'hide' : undefined}
        float="right"
        onClick={() => dispatchMedia({ type: 'nextImage' })}
      >
        <NextIcon />
      </NavigationButton>
      <ExitButton
        className={hide ? 'hide' : undefined}
        onClick={() =>
          dispatchMedia({ type: 'setPresenting', presenting: false })
        }
      >
        <ExitIcon />
      </ExitButton>
    </StyledOverlayContainer>
  )
}

export default PresentNavigationOverlay
