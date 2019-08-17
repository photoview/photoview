import React, { useState } from 'react'
import PropTypes from 'prop-types'
import RouterPropTypes from 'react-router-prop-types'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import Layout from '../../Layout'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'

const AlbumSharePage = ({ album }) => {
  const [activeIndex, setActiveIndex] = useState(-1)
  const [presenting, setPresenting] = useState(false)

  return (
    <Layout>
      <h1>{album.title}</h1>
      <PhotoGallery
        photos={album.photos}
        loading={false}
        activeIndex={activeIndex}
        presenting={presenting}
        onSelectImage={index => {
          setActiveIndex(index)
        }}
        setPresenting={setPresenting}
        nextImage={() => {
          setActiveIndex((activeIndex + 1) % album.photos.length)
        }}
        previousImage={() => {
          setActiveIndex(
            activeIndex < 1 ? album.photos.length - 1 : activeIndex - 1
          )
        }}
      />
    </Layout>
  )
}

AlbumSharePage.propTypes = {
  album: PropTypes.object.isRequired,
}

export default AlbumSharePage
