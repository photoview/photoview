import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'

import ExitIcon from './icons/exit.svg'
import NextIcon from './icons/next.svg'
import PrevIcon from './icons/previous.svg'

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
  cursor: pointer;
  position: absolute;

  & svg {
    width: 32px;
    height: 32px;
  }

  & svg path {
    stroke: rgba(255, 255, 255, 0.4);
    transition: stroke 80ms;
  }

  &:hover svg path {
    stroke: rgba(255, 255, 255, 0.9);
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
  nextImage,
  previousImage,
  setPresenting,
}) => (
  <StyledOverlayContainer>
    <NavigationButton float="left" onClick={() => previousImage()}>
      <PrevIcon />
    </NavigationButton>
    <NavigationButton float="right" onClick={() => nextImage()}>
      <NextIcon />
    </NavigationButton>
    <ExitButton onClick={() => setPresenting(false)}>
      <ExitIcon />
    </ExitButton>
  </StyledOverlayContainer>
)

PresentNavigationOverlay.propTypes = {
  nextImage: PropTypes.func.isRequired,
  previousImage: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
}

export default PresentNavigationOverlay
