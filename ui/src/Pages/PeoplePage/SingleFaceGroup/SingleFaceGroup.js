import PropTypes from 'prop-types'
import React, { useState } from 'react'
import PhotoGallery from '../../../components/photoGallery/PhotoGallery'
import { ProtectedImage } from '../../../components/photoGallery/ProtectedMedia'
import FaceGroupTitle from './FaceGroupTitle'

const ImageFace = ({ imageFace }) => {
  return (
    <div>
      Image face: {imageFace.id}
      <ProtectedImage src={imageFace.media.thumbnail.url} />
    </div>
  )
}

ImageFace.propTypes = {
  imageFace: PropTypes.object.isRequired,
}

const SingleFaceGroup = ({ faceGroup }) => {
  const [presenting, setPresenting] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)

  let mediaGallery = null
  if (faceGroup) {
    const media = faceGroup.imageFaces.map(x => x.media)

    const nextImage = () =>
      setActiveIndex(i => Math.min(i + 1, media.length - 1))

    const previousImage = () => setActiveIndex(i => Math.max(i - 1, 0))

    mediaGallery = (
      <div>
        <PhotoGallery
          media={media}
          loading={false}
          presenting={presenting}
          setPresenting={setPresenting}
          onSelectImage={setActiveIndex}
          activeIndex={activeIndex}
          nextImage={nextImage}
          previousImage={previousImage}
        />
      </div>
    )
  }

  return (
    <div>
      <FaceGroupTitle faceGroup={faceGroup} />
      {mediaGallery}
    </div>
  )
}

SingleFaceGroup.propTypes = {
  faceGroup: PropTypes.object,
}

export default SingleFaceGroup
