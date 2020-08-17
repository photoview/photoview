import React, { useState, useRef, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import debounce from '../../../debounce'

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

const NavigationButton = styled(OverlayButton)`
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

const PresentNavigationOverlay = ({
  children,
  nextImage,
  previousImage,
  setPresenting,
}) => {
  const [hide, setHide] = useState(true)
  const onMouseMove = useRef(null)

  useEffect(() => {
    console.log('Setup mouse move')
    onMouseMove.current = debounce(
      () => {
        setHide(hide => !hide)
      },
      2000,
      true
    )

    return () => {
      onMouseMove.current.cancel()
    }
  }, [])

  return (
    <StyledOverlayContainer
      onMouseMove={() => {
        onMouseMove.current()
      }}
    >
      {children}
      <NavigationButton
        className={hide && 'hide'}
        float="left"
        onClick={() => previousImage()}
      >
        <PrevIcon />
      </NavigationButton>
      <NavigationButton
        className={hide && 'hide'}
        float="right"
        onClick={() => nextImage()}
      >
        <NextIcon />
      </NavigationButton>
      <ExitButton
        className={hide && 'hide'}
        onClick={() => setPresenting(false)}
      >
        <ExitIcon />
      </ExitButton>
    </StyledOverlayContainer>
  )
}

PresentNavigationOverlay.propTypes = {
  children: PropTypes.element,
  nextImage: PropTypes.func.isRequired,
  previousImage: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
}

export default PresentNavigationOverlay
