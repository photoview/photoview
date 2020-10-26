import React from 'react'
import Layout from '../../Layout'
import PropTypes from 'prop-types'
import GalleryGroups from '../../components/photoGallery/GalleryGroups'

const PhotosPage = ({ match }) => {
  return (
    <>
      <Layout title="Photos">
        <GalleryGroups subPage={match.params.subPage} />
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
