import { gql } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useLazyQuery } from '@apollo/client'
import PresentView from '../../components/photoGallery/presentView/PresentView'
import type mapboxgl from 'mapbox-gl'
import { PresentMarker } from './PlacesPage'

const QUERY_MEDIA = gql`
  query placePageQueryMedia($mediaIDs: [ID!]!) {
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

const getMediaFromMarker = (map: mapboxgl.Map, presentMarker: PresentMarker) =>
  new Promise<MediaMarker[]>((resolve, reject) => {
    const { cluster, id } = presentMarker

    if (cluster) {
      const mediaSource = map.getSource('media') as mapboxgl.GeoJSONSource

      mediaSource.getClusterLeaves(id, 1000, 0, (error, features) => {
        if (error) {
          reject(error)
          return
        }

        const media = features.map(feat => feat.properties) as MediaMarker[]
        resolve(media)
      })
    } else {
      const features = map.querySourceFeatures('media')
      const media = features.find(f => f.properties?.media_id == id)
        ?.properties as MediaMarker | undefined

      if (media === undefined) {
        reject('ERROR: media is undefined')
        return
      }

      resolve([media])
    }
  })

export interface MediaMarker {
  id: number
  thumbnail: string
  cluster: boolean
  point_count_abbreviated: number
  cluster_id: number
  media_id: number
}

type MapPresetMarkerProps = {
  map: mapboxgl.Map | null
  presentMarker: PresentMarker | null
  setPresentMarker: React.Dispatch<React.SetStateAction<PresentMarker | null>>
}

const MapPresentMarker = ({
  map,
  presentMarker,
  setPresentMarker,
}: MapPresetMarkerProps) => {
  const [mediaMarkers, setMediaMarkers] = useState<MediaMarker[] | null>(null)
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

  if (
    presentMarker == null ||
    map == null ||
    mediaMarkers == null ||
    loadedMedia == null
  ) {
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

export default MapPresentMarker
