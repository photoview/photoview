import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { ProtectedImage, ProtectedVideo } from '../ProtectedMedia'

const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

const StyledVideo = styled(ProtectedVideo)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
`

const PresentMedia = ({ media, imageLoaded, ...otherProps }) => {
  if (media.type == 'photo') {
    return (
      <div {...otherProps}>
        <StyledPhoto src={media.thumbnail.url} />
        <StyledPhoto
          style={{ display: 'none' }}
          src={media.highRes.url}
          onLoad={e => {
            e.target.style.display = 'initial'
            imageLoaded && imageLoaded()
          }}
        />
      </div>
    )
  }

  if (media.type == 'video') {
    return <StyledVideo media={media} />
  }

  throw new Error(`Unknown media type '${media.type}'`)
}

PresentMedia.propTypes = {
  media: PropTypes.object.isRequired,
  imageLoaded: PropTypes.func,
}

export default PresentMedia
