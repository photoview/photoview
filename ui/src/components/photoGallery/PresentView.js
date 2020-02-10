import React from 'react'
import PropTypes from 'prop-types'
import styled, { createGlobalStyle } from 'styled-components'
import ProtectedImage from './ProtectedImage'

export const PresentContainer = ({ children, ...otherProps }) => {
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

  return (
    <StyledContainer {...otherProps}>
      <PreventScroll />
      {children}
    </StyledContainer>
  )
}

PresentContainer.propTypes = {
  children: PropTypes.any,
}

const PreventScroll = createGlobalStyle`
  body {
    /* height: 100vh !important; */
    overflow: hidden;
  }
`

const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

export const PresentPhoto = ({ photo, imageLoaded, ...otherProps }) => {
  return (
    <div {...otherProps}>
      <StyledPhoto src={photo.thumbnail.url} />
      <StyledPhoto
        style={{ display: 'none' }}
        src={photo.highRes.url}
        onLoad={e => {
          e.target.style.display = 'initial'
          imageLoaded && imageLoaded()
        }}
      />
    </div>
  )
}

PresentPhoto.propTypes = {
  photo: PropTypes.object,
  imageLoaded: PropTypes.func,
}
