import React, { useEffect, useState, useRef } from 'react'
import PropTypes from 'prop-types'
import gql from 'graphql-tag'
import { useLazyQuery } from 'react-apollo'
import PresentView from '../../components/photoGallery/presentView/PresentView'

const QUERY_MEDIA = gql`
  query placePageQueryMedia($mediaID: Int!) {
    media(id: $mediaID) {
      id
      title
      thumbnail {
        url
        width
        height
      }
      highRes {
        url
        width
        height
      }
      videoWeb {
        url
        width
        height
      }
      type
    }
  }
`

const getMediaFromMarker = (map, presentMarker) =>
  new Promise((resolve, reject) => {
    const { cluster, id } = presentMarker

    if (cluster) {
      map
        .getSource('media')
        .getClusterLeaves(id, 1000, 0, (error, features) => {
          if (error) {
            reject(error)
            return
          }

          const media = features.map(feat => feat.properties)
          resolve(media)
        })
    } else {
      const features = map.querySourceFeatures('media')
      const media = features.find(f => f.properties.media_id == id).properties
      resolve([media])
    }
  })

const MapPresentMarker = ({ map, presentMarker, setPresentMarker }) => {
  const [media, setMedia] = useState(null)
  const [currentIndex, setCurrentIndex] = useState(0)

  const previousLoadedMedia = useRef(null)
  const [loadMedia, { data: loadedMedia }] = useLazyQuery(QUERY_MEDIA, {
    onCompleted(data) {
      previousLoadedMedia.current = data
    },
  })

  useEffect(() => {
    if (presentMarker == null || map == null) {
      setMedia(null)
      return
    }

    getMediaFromMarker(map, presentMarker).then(setMedia)
  }, [presentMarker])

  useEffect(() => {
    if (!media) return

    setCurrentIndex(0)
    loadMedia({
      variables: {
        mediaID: media[0].media_id,
      },
    })
  }, [media])

  useEffect(() => {
    if (!media) return

    console.log('Current index change', currentIndex, media)

    loadMedia({
      variables: {
        mediaID: media[currentIndex].media_id,
      },
    })
  }, [currentIndex])

  if (presentMarker == null || map == null) {
    return null
  }

  if (loadedMedia == null && previousLoadedMedia.current == null) {
    return null
  }

  const displayMedia = loadedMedia
    ? loadedMedia.media
    : previousLoadedMedia.current.media

  console.log('diaplay media', displayMedia)

  return (
    <PresentView
      media={displayMedia}
      nextImage={() => {
        setCurrentIndex(i => Math.min(media.length - 1, i + 1))
      }}
      previousImage={() => {
        setCurrentIndex(i => Math.max(0, i - 1))
      }}
      setPresenting={presenting => {
        if (!presenting) {
          previousLoadedMedia.current = null
          setPresentMarker(null)
        }
      }}
    />
  )
}

MapPresentMarker.propTypes = {
  map: PropTypes.object,
  presentMarker: PropTypes.object,
  setPresentMarker: PropTypes.func.isRequired,
}

export default MapPresentMarker
