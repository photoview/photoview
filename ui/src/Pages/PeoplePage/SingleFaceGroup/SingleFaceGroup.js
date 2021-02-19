import PropTypes from 'prop-types'
import React from 'react'
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
  let mediaGallery = null
  if (faceGroup) {
    const media = faceGroup.imageFaces.map(x => x.media)
    mediaGallery = (
      <div>
        <PhotoGallery
          media={media}
          loading={false}
          setPresenting={() => {}}
          onSelectImage={() => {}}
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
