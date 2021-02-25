import { gql, useQuery } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useState } from 'react'
import PaginateLoader from '../../../components/PaginateLoader'
import PhotoGallery from '../../../components/photoGallery/PhotoGallery'
import useScrollPagination from '../../../hooks/useScrollPagination'
import FaceGroupTitle from './FaceGroupTitle'

export const SINGLE_FACE_GROUP = gql`
  query singleFaceGroup($id: ID!, $limit: Int!, $offset: Int!) {
    faceGroup(id: $id) {
      id
      label
      imageFaces(paginate: { limit: $limit, offset: $offset }) {
        id
        rectangle {
          minX
          maxX
          minY
          maxY
        }
        media {
          id
          type
          title
          thumbnail {
            url
            width
            height
          }
          highRes {
            url
          }
          favorite
        }
      }
    }
  }
`

const SingleFaceGroup = ({ faceGroupID }) => {
  const { data, error, loading, fetchMore } = useQuery(SINGLE_FACE_GROUP, {
    variables: {
      limit: 2,
      offset: 0,
      id: faceGroupID,
    },
  })
  const [presenting, setPresenting] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)

  const { containerElem, finished: finishedLoadingMore } = useScrollPagination({
    loading,
    fetchMore,
    data,
    getItems: data => data.faceGroup.imageFaces,
  })

  const faceGroup = data?.faceGroup

  if (error) {
    return <div>{error.message}</div>
  }

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
        <PaginateLoader
          active={!finishedLoadingMore && !loading}
          text="Loading more photos"
        />
      </div>
    )
  }

  return (
    <div ref={containerElem}>
      <FaceGroupTitle faceGroup={faceGroup} />
      {mediaGallery}
    </div>
  )
}

SingleFaceGroup.propTypes = {
  faceGroupID: PropTypes.string,
}

export default SingleFaceGroup
