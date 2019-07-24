import React from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { useSpring, animated } from 'react-spring'
import LazyLoad from 'react-lazyload'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

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

class Photo extends React.Component {
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
    `}
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

class PhotoGallery extends React.Component {
  constructor(props) {
    super(props)

    // this.keyDownEvent = e => {
    //   if (!this.props.onSelectImage) {
    //     return
    //   }

    //   const activeImage = this.state.activeImage
    //   if (activeImage != -1) {
    //     if (e.key == 'ArrowRight') {
    //       this.setActiveImage((activeImage + 1) % this.photoAmount)
    //     }

    //     if (e.key == 'ArrowLeft') {
    //       if (activeImage <= 0) {
    //         this.setActiveImage(this.photoAmount - 1)
    //       } else {
    //         this.setActiveImage(activeImage - 1)
    //       }
    //     }
    //   }
    // }
  }

  // componentDidMount() {
  //   document.addEventListener('keydown', this.keyDownEvent)
  // }

  // componentWillUnmount() {
  //   document.removeEventListener('keydown', this.keyDownEvent)
  // }

  render() {
    const { activeIndex = -1, photos, loading, onSelectImage } = this.props
    let photoElements = null
    if (photos) {
      photoElements = photos.map((photo, index) => {
        const active = activeIndex == index

        let minWidth = 100
        if (photo.thumbnail) {
          minWidth = Math.floor(
            (photo.thumbnail.width / photo.thumbnail.height) * 200
          )
        }

        return (
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
            <Photo src={photo.thumbnail && photo.thumbnail.url} />
            <PhotoOverlay active={active} />
          </PhotoContainer>
        )
      })
    }

    return (
      <div>
        <Gallery>
          <Loader active={loading}>Loading images</Loader>
          {photoElements}
          <PhotoFiller />
        </Gallery>
      </div>
    )
  }
}
export default PhotoGallery
