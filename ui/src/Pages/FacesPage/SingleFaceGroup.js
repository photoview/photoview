import React from 'react'
import PropTypes from 'prop-types'
import { ProtectedImage } from '../../components/photoGallery/ProtectedMedia'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'

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
  if (!faceGroup) {
    return null
  }

  // const images = faceGroup.imageFaces.map(imgFace => (
  //   <ImageFace key={imgFace.id} imageFace={imgFace} />
  // ))

  const media = faceGroup.imageFaces.map(x => x.media)

  return (
    <div>
      Face group: {faceGroup.id}
      <div>
        <PhotoGallery media={media} loading={false} setPresenting={() => {}} />
      </div>
    </div>
  )
}

SingleFaceGroup.propTypes = {
  faceGroup: PropTypes.object,
}

export default SingleFaceGroup
