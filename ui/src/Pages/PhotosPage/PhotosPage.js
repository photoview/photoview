import React from 'react'
import Layout from '../../Layout'
import PropTypes from 'prop-types'
import TimelineGallery from '../../components/timelineGallery/TimelineGallery'

const PhotosPage = () => {
  return (
    <>
      <Layout title="Photos">
        <TimelineGallery />
      </Layout>
    </>
  )
}

PhotosPage.propTypes = {
  match: PropTypes.shape({
    params: PropTypes.shape({
      subPage: PropTypes.string,
    }),
  }),
}

export default PhotosPage
