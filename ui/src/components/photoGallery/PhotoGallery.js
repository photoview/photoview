import React from 'react'
import styled from 'styled-components'
import { Loader } from 'semantic-ui-react'
import { Photo } from './Photo'
import PresentView from './PresentView'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

const presentIdFromHash = hash => {
  let match = hash.match(/present=([a-z0-9\-]+)/)
  return match && match[1]
}

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
          <Photo
            key={photo.id}
            photo={photo}
            onSelectImage={onSelectImage}
            minWidth={minWidth}
            index={index}
            active={active}
          />
        )
      })
    }

    console.log(presentIdFromHash(location.hash))

    return (
      <div>
        <Gallery>
          <Loader active={loading}>Loading images</Loader>
          {photoElements}
          <PhotoFiller />
        </Gallery>
        <PresentView image={presentIdFromHash(location.hash)} />
      </div>
    )
  }
}
export default PhotoGallery
