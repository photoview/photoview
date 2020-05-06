import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'

const StyledOverlayContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const PresentNavigationOverlay = ({
  nextImage,
  previousImage,
  setPresenting,
}) => (
  <StyledOverlayContainer>
    <button onClick={() => setPresenting(false)}>Exit</button>
    <button onClick={() => previousImage()}>Previous</button>
    <button onClick={() => nextImage()}>Next</button>
  </StyledOverlayContainer>
)

PresentNavigationOverlay.propTypes = {
  nextImage: PropTypes.func.isRequired,
  previousImage: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
}

export default PresentNavigationOverlay
