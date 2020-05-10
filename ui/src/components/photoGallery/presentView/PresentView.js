import PropTypes from 'prop-types'
import React from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentPhoto from './PresentPhoto'

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
  photo,
  imageLoaded,
  nextImage,
  previousImage,
  setPresenting,
}) => (
  <StyledContainer {...className}>
    <PreventScroll />
    <PresentPhoto photo={photo} imageLoaded={imageLoaded} />
    <PresentNavigationOverlay
      {...{ nextImage, previousImage, setPresenting }}
    />
  </StyledContainer>
)

PresentView.propTypes = {
  photo: PropTypes.object.isRequired,
  imageLoaded: PropTypes.func,
  className: PropTypes.string,
  nextImage: PropTypes.func.isRequired,
  previousImage: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
}

export default PresentView
