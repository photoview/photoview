import React from 'react'
import styled from 'styled-components'
import { useSpring, animated } from 'react-spring'
import LazyLoad from 'react-lazyload'
import { Icon } from 'semantic-ui-react'

const PhotoContainer = styled.div`
  flex-grow: 1;
  height: 200px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`

const PhotoImg = photoProps => {
  const StyledPhoto = styled(animated.img)`
    height: 200px;
    min-width: 100%;
    position: relative;
    object-fit: cover;
  `

  const [props, set, stop] = useSpring(() => ({ opacity: 0 }))

  return (
    <StyledPhoto
      {...photoProps}
      style={props}
      onLoad={() => {
        set({ opacity: 1 })
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
      <LazyLoad>
        <PhotoImg {...this.props} />
      </LazyLoad>
    )
  }
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
