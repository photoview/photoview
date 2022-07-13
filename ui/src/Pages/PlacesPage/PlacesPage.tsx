import { gql, useQuery } from '@apollo/client'
import type mapboxgl from 'mapbox-gl'
import React, { useReducer } from 'react'
import { Helmet } from 'react-helmet'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import Layout from '../../components/layout/Layout'
import { registerMediaMarkers } from '../../components/mapbox/mapboxHelperFunctions'
import useMapboxMap from '../../components/mapbox/MapboxMap'
import { urlPresentModeSetupHook } from '../../components/photoGallery/mediaGalleryReducer'
import MapPresentMarker from './MapPresentMarker'
import { PlacesAction, placesReducer } from './placesReducer'
import { mediaGeoJson } from './__generated__/mediaGeoJson'

const MapWrapper = styled.div`
  width: 100%;
  height: calc(100vh - 120px);
`

const MAPBOX_DATA_QUERY = gql`
  query mediaGeoJson {
    myMediaGeoJson
  }
`

export type PresentMarker = {
  id: number | string
  cluster: boolean
}

const MapPage = () => {
  const { t } = useTranslation()

  const { data: mapboxData } = useQuery<mediaGeoJson>(MAPBOX_DATA_QUERY, {
    fetchPolicy: 'cache-first',
  })

  const [markerMediaState, dispatchMarkerMedia] = useReducer(placesReducer, {
    presenting: false,
    activeIndex: -1,
    media: [],
  })

  const { mapContainer, mapboxMap, mapboxToken } = useMapboxMap({
    configureMapbox: configureMapbox({ mapboxData, dispatchMarkerMedia }),
  })

  urlPresentModeSetupHook({
    dispatchMedia: dispatchMarkerMedia,
    openPresentMode: event => {
      dispatchMarkerMedia({
        type: 'openPresentMode',
        activeIndex: event.state.activeIndex,
      })
    },
  })

  if (mapboxData && mapboxToken == null) {
    return (
      <Layout title={t('places_page.title', 'Places')}>
        <h1>Mapbox token is not set</h1>
        <p>
          To use map related features a mapbox token is needed.
          <br /> A mapbox token can be created for free at{' '}
          <a href="https://account.mapbox.com/access-tokens/">mapbox.com</a>.
        </p>
        <p>
          Make sure the access token is added as the MAPBOX_TOKEN environment
          variable.
        </p>
      </Layout>
    )
  }

  return (
    <Layout title="Places">
      <Helmet>
        {/* <link rel="stylesheet" href="/mapbox-gl.css" /> */}
        {/* <style type="text/css">{mapboxStyles}</style> */}
      </Helmet>
      <MapWrapper>{mapContainer}</MapWrapper>
      <MapPresentMarker
        map={mapboxMap}
        markerMediaState={markerMediaState}
        dispatchMarkerMedia={dispatchMarkerMedia}
      />
    </Layout>
  )
}

const configureMapbox =
  ({
    mapboxData,
    dispatchMarkerMedia,
  }: {
    mapboxData?: mediaGeoJson
    dispatchMarkerMedia: React.Dispatch<PlacesAction>
  }) =>
  (map: mapboxgl.Map, mapboxLibrary: typeof mapboxgl) => {
    // Add map navigation control
    map.addControl(new mapboxLibrary.NavigationControl())

    map.on('load', () => {
      if (map == null) {
        console.error('ERROR: map is null')
        return
      }

      map.addSource('media', {
        type: 'geojson',
        data: mapboxData?.myMediaGeoJson as never,
        cluster: true,
        clusterRadius: 50,
        clusterProperties: {
          thumbnail: ['coalesce', ['get', 'thumbnail'], false],
        },
      })

      // Add dummy layer for features to be queryable
      map.addLayer({
        id: 'media-points',
        type: 'circle',
        source: 'media',
        filter: ['!', true],
      })

      registerMediaMarkers({
        map: map,
        mapboxLibrary,
        dispatchMarkerMedia,
      })
    })
  }

export default MapPage
