import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { useSpring, animated } from 'react-spring'
import LazyLoad from 'react-lazyload'
import { Icon } from 'semantic-ui-react'
import ProtectedImage from './ProtectedImage'

const PhotoContainer = styled.div`
  flex-grow: 1;
  height: 200px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`

const StyledPhoto = styled(ProtectedImage)`
  height: 200px;
  min-width: 100%;
  position: relative;
  object-fit: cover;
  opacity: ${props => (props.loaded ? 1 : 0)};

  transition: opacity 300ms;
`

const PhotoImg = photoProps => {
  const [loaded, setLoaded] = useState(false)

  return (
    <StyledPhoto
      {...photoProps}
      loaded={loaded}
      onLoad={() => {
        setLoaded(true)
      }}
    />
  )
}

class LazyPhoto extends React.Component {
  shouldComponentUpdate(nextProps) {
    return nextProps.src != this.props.src
  }

  render() {
    return (
      <LazyLoad scrollContainer="#layout-content">
        <PhotoImg {...this.props} />
      </LazyLoad>
    )
  }
}

LazyPhoto.propTypes = {
  src: PropTypes.string,
}

const PhotoOverlay = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;

  ${props =>
    props.active &&
    `
      border: 4px solid rgba(65, 131, 196, 0.6);

      & ${HoverIcon} {
        top: -4px !important;
        left: -4px !important;
      }
    `}
`

const HoverIcon = styled(Icon)`
  font-size: 1.5em !important;
  margin: 160px 0 0 10px !important;
  color: white !important;
  text-shadow: 0 0 4px black;
  opacity: 0 !important;
  position: relative;

  border-radius: 50%;
  width: 34px !important;
  height: 34px !important;
  padding-top: 8px;

  ${PhotoContainer}:hover & {
    opacity: 1 !important;
  }

  &:hover {
    background-color: rgba(255, 255, 255, 0.4);
  }

  transition: opacity 100ms, background-color 100ms;
`

export const Photo = ({
  photo,
  onSelectImage,
  minWidth,
  index,
  active,
  setPresenting,
}) => (
  <PhotoContainer
    key={photo.id}
    style={{
      cursor: onSelectImage ? 'pointer' : null,
      minWidth: `${minWidth}px`,
    }}
    onClick={() => {
      onSelectImage && onSelectImage(index)
    }}
  >
    <LazyPhoto src={photo.thumbnail && photo.thumbnail.url} />
    <PhotoOverlay active={active}>
      <HoverIcon
        name="expand"
        onClick={() => {
          // window.location.hash = `present=${photo.id}`
          setPresenting(true)
        }}
      />
      <HoverIcon name="heart outline" />
    </PhotoOverlay>
  </PhotoContainer>
)

Photo.propTypes = {
  photo: PropTypes.object.isRequired,
  onSelectImage: PropTypes.func,
  minWidth: PropTypes.number.isRequired,
  index: PropTypes.number.isRequired,
  active: PropTypes.bool.isRequired,
  setPresenting: PropTypes.func.isRequired,
}
