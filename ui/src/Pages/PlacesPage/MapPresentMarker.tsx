import { gql } from '@apollo/client'
import React, { useEffect } from 'react'
import { useLazyQuery } from '@apollo/client'
import PresentView from '../../components/photoGallery/presentView/PresentView'
import type mapboxgl from 'mapbox-gl'
import { PresentMarker } from './PlacesPage'
import {
  placePageQueryMedia,
  placePageQueryMediaVariables,
} from './__generated__/placePageQueryMedia'
import { PlacesAction, PlacesState } from './placesReducer'

const QUERY_MEDIA = gql`
  query placePageQueryMedia($mediaIDs: [ID!]!) {
    mediaList(ids: $mediaIDs) {
      id
      title
      blurhash
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

      mediaSource.getClusterLeaves(id as number, 1000, 0, (error, features) => {
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
  cluster_id: string
  media_id: string
}

type MapPresetMarkerProps = {
  map: mapboxgl.Map | null
  markerMediaState: PlacesState
  dispatchMarkerMedia: React.Dispatch<PlacesAction>
}

/**
 * Full-screen present-view that works with PlacesState
 */
const MapPresentMarker = ({
  map,
  markerMediaState,
  dispatchMarkerMedia,
}: MapPresetMarkerProps) => {
  const [loadMedia, { data: loadedMedia }] = useLazyQuery<
    placePageQueryMedia,
    placePageQueryMediaVariables
  >(QUERY_MEDIA)

  useEffect(() => {
    const presentMarker = markerMediaState.presentMarker
    if (presentMarker == null || map == null) {
      dispatchMarkerMedia({
        type: 'closePresentMode',
      })
      return
    }

    getMediaFromMarker(map, presentMarker).then(mediaMarkers => {
      loadMedia({
        variables: {
          mediaIDs: mediaMarkers.map(x => x.media_id),
        },
      })
    })
  }, [markerMediaState.presentMarker])

  useEffect(() => {
    const mediaList = loadedMedia?.mediaList || []
    dispatchMarkerMedia({
      type: 'replaceMedia',
      media: mediaList,
    })
    if (mediaList.length > 0) {
      dispatchMarkerMedia({
        type: 'openPresentMode',
        activeIndex: 0,
      })
    }
  }, [loadedMedia])

  if (markerMediaState.presenting) {
    return (
      <PresentView
        activeMedia={markerMediaState.media[markerMediaState.activeIndex]}
        dispatchMedia={dispatchMarkerMedia}
        disableSaveCloseInHistory={true}
      />
    )
  } else {
    return null
  }
}

export default MapPresentMarker
