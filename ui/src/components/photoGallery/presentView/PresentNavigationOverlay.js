import React from 'react'
import styled from 'styled-components'

const StyledOverlayContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const PresentNavigationOverlay = () => (
  <StyledOverlayContainer>
    <button>Previous</button>
    <button>Next</button>
  </StyledOverlayContainer>
)

export default PresentNavigationOverlay
