import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import { presentIndexFromHash } from '../photoGallery/PhotoGallery'
import Layout from '../../Layout'
import AlbumTitle from '../AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'

const AlbumGallery = ({ album, loading = false, customAlbumLink }) => {
  const [activeImage, setActiveImage] = useState(-1)
  const [presenting, setPresenting] = useState(false)

  const nextImage = () => {
    setActiveImage((activeImage + 1) % album.photos.length)
  }

  const previousImage = () => {
    if (activeImage <= 0) {
      setActiveImage(album.photos.length - 1)
    } else {
      setActiveImage(activeImage - 1)
    }
  }

  // Setup
  useEffect(() => {
    const presentIndex = presentIndexFromHash(document.location.hash)
    if (presentIndex) {
      setActiveImage(presentIndex)
      setPresenting(true)
    }
  }, [])

  // On update
  useEffect(() => {
    if (presenting) {
      window.history.replaceState(
        null,
        null,
        document.location.pathname + '#' + `present=${activeImage}`
      )
    } else if (presentIndexFromHash(document.location.hash)) {
      window.history.replaceState(
        null,
        null,
        document.location.pathname.split('#')[0]
      )
    }
  })

  useEffect(() => {
    setActiveImage(-1)
  }, [album])

  let subAlbumElement = null

  if (album) {
    if (album.subAlbums.length > 0) {
      subAlbumElement = (
        <AlbumBoxes
          loading={loading}
          albums={album.subAlbums}
          getCustomLink={customAlbumLink}
        />
      )
    }
  }

  return (
    <Layout>
      <AlbumTitle album={album} disableLink />
      {subAlbumElement}
      {album && album.subAlbums.length > 0 && <h2>Images</h2>}
      <PhotoGallery
        loading={loading}
        photos={album && album.photos}
        activeIndex={activeImage}
        presenting={presenting}
        onSelectImage={index => {
          setActiveImage(index)
        }}
        setPresenting={setPresenting}
        nextImage={nextImage}
        previousImage={previousImage}
      />
    </Layout>
  )
}

AlbumGallery.propTypes = {
  album: PropTypes.object,
  loading: PropTypes.bool,
  customAlbumLink: PropTypes.func,
}

export default AlbumGallery
