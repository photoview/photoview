import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import Layout from '../../Layout'
import AlbumTitle from '../AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'

const AlbumGallery = ({ album, loading = false, customAlbumLink }) => {
  const [activeImage, setActiveImage] = useState(-1)
  const [presenting, setPresenting] = useState(false)

  const setPresentingWithHistory = presenting => {
    setPresenting(presenting)
    if (presenting) {
      history.pushState({ presenting: true }, '')
    } else {
      history.back()
    }
  }

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

  useEffect(() => {
    const updatePresenting = event => {
      setPresenting(event.state.presenting)
    }
    window.addEventListener('popstate', updatePresenting)

    return () => {
      window.removeEventListener('popstate', updatePresenting)
    }
  }, [activeImage])

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
  } else {
    subAlbumElement = <AlbumBoxes loading={loading} />
  }

  return (
    <Layout title={album ? album.title : 'Loading album'}>
      <AlbumTitle album={album} disableLink />
      {subAlbumElement}
      {
        <h2
          style={{
            opacity: loading ? 0 : 1,
            display: album && album.subAlbums.length > 0 ? 'block' : 'none',
          }}
        >
          Images
        </h2>
      }
      <PhotoGallery
        loading={loading}
        photos={album && album.photos}
        activeIndex={activeImage}
        presenting={presenting}
        onSelectImage={index => {
          setActiveImage(index)
        }}
        setPresenting={setPresentingWithHistory}
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
