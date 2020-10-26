import { gql } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useEffect, useState } from 'react'
import { useLazyQuery } from '@apollo/client'
import PresentView from '../../components/photoGallery/presentView/PresentView'

const QUERY_MEDIA = gql`
  query placePageQueryMedia($mediaIDs: [Int!]!) {
    mediaList(ids: $mediaIDs) {
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
  const [mediaMarkers, setMediaMarkers] = useState(null)
  const [currentIndex, setCurrentIndex] = useState(0)

  const [loadMedia, { data: loadedMedia }] = useLazyQuery(QUERY_MEDIA)

  useEffect(() => {
    if (presentMarker == null || map == null) {
      setMediaMarkers(null)
      return
    }

    getMediaFromMarker(map, presentMarker).then(setMediaMarkers)
  }, [presentMarker])

  useEffect(() => {
    if (!mediaMarkers) return

    setCurrentIndex(0)
    loadMedia({
      variables: {
        mediaIDs: mediaMarkers.map(x => x.media_id),
      },
    })
  }, [mediaMarkers])

  if (presentMarker == null || map == null) {
    return null
  }

  if (loadedMedia == null) {
    return null
  }

  return (
    <PresentView
      media={loadedMedia.mediaList[currentIndex]}
      nextImage={() => {
        setCurrentIndex(i => Math.min(mediaMarkers.length - 1, i + 1))
      }}
      previousImage={() => {
        setCurrentIndex(i => Math.max(0, i - 1))
      }}
      setPresenting={presenting => {
        if (!presenting) {
          setCurrentIndex(0)
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
