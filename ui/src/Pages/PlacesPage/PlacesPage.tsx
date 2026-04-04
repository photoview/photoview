import { gql, useQuery } from '@apollo/client'
import type maplibregl from 'maplibre-gl'
import React, { useCallback, useEffect, useReducer, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import Layout from '../../components/layout/Layout'
import { findBestView, extractCoordsFromGeoJson } from '../../components/maplibre/globeCenter'
import { registerMediaMarkers } from '../../components/maplibre/maplibreHelperFunctions'
import useMaplibreMap from '../../components/maplibre/MaplibreMap'
import { urlPresentModeSetupHook } from '../../components/photoGallery/mediaGalleryReducer'
import MapPresentMarker from './MapPresentMarker'
import { PlacesAction, placesReducer } from './placesReducer'
import { mediaGeoJson } from './__generated__/mediaGeoJson'

const MapWrapper = styled.div`
  width: 100%;
  height: calc(100vh - 120px);
  position: relative;
`

const Overlay = styled.div`
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.6);
  z-index: 1;

  .dark & {
    background: rgba(0, 0, 0, 0.5);
  }
`

const MAP_DATA_QUERY = gql`
  query mediaGeoJson {
    myMediaGeoJson
  }
`

const EMPTY_FEATURE_COLLECTION: GeoJSON.FeatureCollection = {
  type: 'FeatureCollection',
  features: [],
}

export type PresentMarker = {
  id: number | string
  cluster: boolean
}

const MapPage = () => {
  const { t, i18n } = useTranslation()

  const { data: mapData, loading, error } = useQuery<mediaGeoJson>(MAP_DATA_QUERY, {
    fetchPolicy: 'cache-first',
  })

  const hasCentered = useRef(false)

  const [markerMediaState, dispatchMarkerMedia] = useReducer(placesReducer, {
    presenting: false,
    activeIndex: -1,
    media: [],
  })

  const cleanupMarkersRef = useRef<(() => void) | null>(null)

  const configureMap = useCallback(
    (map: maplibregl.Map, maplibreLibrary: typeof maplibregl) => {
      map.addControl(new maplibreLibrary.NavigationControl())
      map.addControl(new maplibreLibrary.GlobeControl())
      map.addControl(new maplibreLibrary.GeolocateControl())
    },
    []
  )

  const onStyleLoad = useCallback(
    (map: maplibregl.Map, maplibreLibrary: typeof maplibregl) => {
      cleanupMarkersRef.current?.()
      cleanupMarkersRef.current = null

      if (map.getSource('media')) return

      const data = mapData?.myMediaGeoJson ?? EMPTY_FEATURE_COLLECTION

      map.addSource('media', {
        type: 'geojson',
        data: data as never,
        cluster: true,
        clusterRadius: 50,
        clusterProperties: {
          thumbnail: ['coalesce', ['get', 'thumbnail'], false],
        },
      })

      map.addLayer({
        id: 'media-points',
        type: 'circle',
        source: 'media',
        filter: ['!', true],
      })

      cleanupMarkersRef.current = registerMediaMarkers({
        map,
        maplibreLibrary,
        dispatchMarkerMedia,
      })
    },
    [mapData, dispatchMarkerMedia]
  )

  const { mapContainer, maplibreMap } = useMaplibreMap({
    configureMap,
    onStyleLoad,
    locale: i18n.language,
  })

  // Update GeoJSON data when query resolves after map is ready
  useEffect(() => {
    if (!maplibreMap || !mapData?.myMediaGeoJson) return

    const source = maplibreMap.getSource('media') as maplibregl.GeoJSONSource | undefined
    if (source) {
      source.setData(mapData.myMediaGeoJson as never)
    }

    if (!hasCentered.current) {
      const coords = extractCoordsFromGeoJson(mapData.myMediaGeoJson as GeoJSON.FeatureCollection)
      const { center, zoom } = findBestView(coords)
      maplibreMap.jumpTo({ center: [center.lng, center.lat], zoom })
      hasCentered.current = true
    }
  }, [maplibreMap, mapData])

  urlPresentModeSetupHook({
    dispatchMedia: dispatchMarkerMedia,
    openPresentMode: event => {
      dispatchMarkerMedia({
        type: 'openPresentMode',
        activeIndex: event.state.activeIndex,
      })
    },
  })

  return (
    <Layout title={t('places_page.title', 'Places')}>
      <MapWrapper>
        {mapContainer}
        {loading && (
          <Overlay>
            <span>{t('general.loading.default', 'Loading...')}</span>
          </Overlay>
        )}
        {error && (
          <Overlay>
            <span>{t('general.loading.error.description', 'An error occurred')}: {error.message}</span>
          </Overlay>
        )}
      </MapWrapper>
      <MapPresentMarker
        map={maplibreMap}
        markerMediaState={markerMediaState}
        dispatchMarkerMedia={dispatchMarkerMedia}
      />
    </Layout>
  )
}

export default MapPage
