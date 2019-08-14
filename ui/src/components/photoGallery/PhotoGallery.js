import React from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { Photo } from './Photo'
import PresentView from './PresentView'
import PropTypes from 'prop-types'
import { fetchProtectedImage } from './ProtectedImage'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

export const presentIndexFromHash = hash => {
  let match = hash.match(/present=(\d+)/)
  return match && parseInt(match[1])
}

class PhotoGallery extends React.Component {
  constructor(props) {
    super(props)

    this.keyDownEvent = e => {
      if (!this.props.onSelectImage || this.props.activeIndex == -1) {
        return
      }

      if (e.key == 'ArrowRight') {
        this.props.nextImage && this.props.nextImage()
      }

      if (e.key == 'ArrowLeft') {
        this.props.nextImage && this.props.previousImage()
      }

      if (e.key == 'Escape') {
        this.props.setPresenting(false)
      }
    }

    this.preloadImages = this.preloadImages.bind(this)
  }

  componentDidMount() {
    document.addEventListener('keydown', this.keyDownEvent)
  }

  componentWillUnmount() {
    document.removeEventListener('keydown', this.keyDownEvent)
  }

  preloadImages() {
    async function preloadImage(url) {
      var img = new Image()
      img.src = await fetchProtectedImage(url)
    }

    const { activeIndex = -1, photos } = this.props

    if (activeIndex != -1 && photos) {
      let previousIndex = null
      let nextIndex = null

      if (activeIndex > 0) {
        previousIndex = activeIndex - 1
      } else {
        previousIndex = photos.length - 1
      }

      nextIndex = (activeIndex + 1) % photos.length

      preloadImage(photos[nextIndex].original.url)
      preloadImage(photos[previousIndex].original.url)
    }
  }

  render() {
    const {
      activeIndex = -1,
      photos,
      loading,
      onSelectImage,
      presenting,
    } = this.props

    const activeImage = photos && activeIndex != -1 && photos[activeIndex]

    let photoElements = null
    if (photos) {
      photos.filter(photo => photo.thumbnail)

      photoElements = photos.map((photo, index) => {
        const active = activeIndex == index

        let minWidth = 100
        if (photo.thumbnail) {
          minWidth = Math.floor(
            (photo.thumbnail.width / photo.thumbnail.height) * 200
          )
        }

        return (
          <Photo
            key={photo.id}
            photo={photo}
            onSelectImage={onSelectImage}
            setPresenting={this.props.setPresenting}
            minWidth={minWidth}
            index={index}
            active={active}
          />
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
        <PresentView
          presenting={presenting}
          image={activeImage && activeImage.id}
          thumbnail={activeImage && activeImage.thumbnail.url}
          imageLoaded={this.preloadImages()}
        />
      </div>
    )
  }
}

PhotoGallery.propTypes = {
  loading: PropTypes.bool,
  photos: PropTypes.array,
  activeIndex: PropTypes.number,
  presenting: PropTypes.bool,
  onSelectImage: PropTypes.func,
  setPresenting: PropTypes.func,
  nextImage: PropTypes.func,
  previousImage: PropTypes.func,
}

export default PhotoGallery
