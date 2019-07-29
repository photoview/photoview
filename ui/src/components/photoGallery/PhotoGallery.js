import React from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { Photo } from './Photo'
import PresentView from './PresentView'
import PropTypes from 'prop-types'

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
  }

  componentDidMount() {
    document.addEventListener('keydown', this.keyDownEvent)
  }

  componentWillUnmount() {
    document.removeEventListener('keydown', this.keyDownEvent)
  }

  render() {
    const { activeIndex = -1, photos, loading, onSelectImage } = this.props

    const activeImage = photos && activeIndex != -1 && photos[activeIndex]

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
          presenting={this.props.presenting}
          image={activeImage && activeImage.id}
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
