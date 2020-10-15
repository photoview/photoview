import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import Layout from '../../Layout'
import AlbumTitle from '../AlbumTitle'
import PhotoGallery from '../photoGallery/PhotoGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../AlbumFilter'

const AlbumGallery = ({
  album,
  loading = false,
  customAlbumLink,
  showFilter = false,
  setOnlyFavorites,
  setOrdering,
  onlyFavorites = false,
  onFavorite,
}) => {
  const [imageState, setImageState] = useState({
    activeImage: -1,
    presenting: false,
  })

  const setPresenting = presenting =>
    setImageState(state => ({ ...state, presenting }))

  const setPresentingWithHistory = presenting => {
    setPresenting(presenting)
    if (presenting) {
      history.pushState({ imageState }, '')
    } else {
      history.back()
    }
  }

  const updateHistory = imageState => {
    history.replaceState({ imageState }, '')
    return imageState
  }

  const setActiveImage = activeImage => {
    setImageState(state => updateHistory({ ...state, activeImage }))
  }

  const nextImage = () => {
    setActiveImage((imageState.activeImage + 1) % album.media.length)
  }

  const previousImage = () => {
    if (imageState.activeImage <= 0) {
      setActiveImage(album.media.length - 1)
    } else {
      setActiveImage(imageState.activeImage - 1)
    }
  }

  useEffect(() => {
    const updateImageState = event => {
      // console.log('Getting status from history', event.state)
      setImageState(event.state.imageState)
    }
    window.addEventListener('popstate', updateImageState)

    return () => {
      window.removeEventListener('popstate', updateImageState)
    }
  }, [imageState])

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
      {showFilter && (
        <AlbumFilter
          onlyFavorites={onlyFavorites}
          setOnlyFavorites={setOnlyFavorites}
          setOrdering={setOrdering}
        />
      )}
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
        media={album && album.media}
        activeIndex={imageState.activeImage}
        presenting={imageState.presenting}
        onSelectImage={index => {
          setActiveImage(index)
        }}
        onFavorite={onFavorite}
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
  showFilter: PropTypes.bool,
  setOnlyFavorites: PropTypes.func,
  onlyFavorites: PropTypes.bool,
  onFavorite: PropTypes.func,
  setOrdering: PropTypes.func,
}

export default AlbumGallery
