import React, { useEffect, useRef, useState } from 'react'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import styled from 'styled-components'

import 'mapbox-gl/dist/mapbox-gl.css'

import Layout from '../../Layout'
import { makeUpdateMarkers } from './mapboxHelperFunctions'
import MapPresentMarker from './MapPresentMarker'

const MapWrapper = styled.div`
  width: 100%;
  height: calc(100% - 24px);
`

const MapContainer = styled.div`
  width: 100%;
  height: 100%;
`

const MAPBOX_DATA_QUERY = gql`
  query placePageMapboxToken {
    mapboxToken
    myMediaGeoJson
  }
`

const MapPage = () => {
  const [mapboxLibrary, setMapboxLibrary] = useState(null)
  const [presentMarker, setPresentMarker] = useState(null)
  const mapContainer = useRef()
  const map = useRef()

  const { data: mapboxData } = useQuery(MAPBOX_DATA_QUERY)

  useEffect(() => {
    async function loadMapboxLibrary() {
      const mapbox = await import('mapbox-gl')
      // mapbox.accessToken = <INSERT ACCESS TOKEN>
      setMapboxLibrary(mapbox)
    }
    loadMapboxLibrary()
  }, [])

  useEffect(() => {
    if (
      mapboxLibrary == null ||
      mapContainer.current == null ||
      mapboxData == null ||
      map.current != null
    ) {
      return
    }

    mapboxLibrary.accessToken = mapboxData.mapboxToken

    map.current = new mapboxLibrary.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/streets-v11',
      zoom: 1,
    })

    map.current.on('load', () => {
      console.log(mapboxData.myMediaGeoJson)
      map.current.addSource('media', {
        type: 'geojson',
        data: mapboxData.myMediaGeoJson,
        cluster: true,
        // clusterMaxZoom: 14, // Max zoom to cluster points on
        clusterRadius: 50,
        clusterProperties: {
          thumbnail: ['coalesce', ['get', 'thumbnail'], false],
        },
      })

      // Add dummy layer for features to be queryable
      map.current.addLayer({
        id: 'media-points',
        type: 'circle',
        source: 'media',
        filter: ['!', true],
      })

      const updateMarkers = makeUpdateMarkers({
        map: map.current,
        mapboxLibrary,
        setPresentMarker,
      })

      map.current.on('move', updateMarkers)
      map.current.on('moveend', updateMarkers)
      map.current.on('sourcedata', updateMarkers)
      updateMarkers()
    })
  }, [mapContainer, mapboxLibrary, mapboxData])

  return (
    <Layout>
      <MapWrapper>
        <MapContainer ref={mapContainer}></MapContainer>
      </MapWrapper>
      <MapPresentMarker
        map={map.current}
        presentMarker={presentMarker}
        setPresentMarker={setPresentMarker}
      />
    </Layout>
  )
}

export default MapPage
