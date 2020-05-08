import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import ProtectedImage from '../ProtectedImage'

const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

const PresentPhoto = ({ photo, imageLoaded, ...otherProps }) => (
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

PresentPhoto.propTypes = {
  photo: PropTypes.object.isRequired,
  imageLoaded: PropTypes.func,
}

export default PresentPhoto
