import React, { useState, useRef, useEffect } from 'react'
import { gql, useQuery } from '@apollo/client'
import type mapboxgl from 'mapbox-gl'
import styled from 'styled-components'

import 'mapbox-gl/dist/mapbox-gl.css'
import { mapboxToken } from './__generated__/mapboxToken'

const MAPBOX_TOKEN_QUERY = gql`
  query mapboxToken {
    mapboxToken
    myMediaGeoJson
  }
`

const MapContainer = styled.div`
  width: 100%;
  height: 100%;
`

type MapboxMapProps = {
  configureMapbox(map: mapboxgl.Map, mapboxLibrary: typeof mapboxgl): void
  readonly initialZoom?: number
}

const useMapboxMap = ({ configureMapbox, initialZoom = 1 }: MapboxMapProps) => {
  const [mapboxLibrary, setMapboxLibrary] = useState<typeof mapboxgl>()
  const mapContainer = useRef<HTMLDivElement | null>(null)
  const map = useRef<mapboxgl.Map | null>(null)

  const { data: mapboxData } = useQuery<mapboxToken>(MAPBOX_TOKEN_QUERY, {
    fetchPolicy: 'cache-first',
  })

  useEffect(() => {
    async function loadMapboxLibrary() {
      const mapbox = (await import('mapbox-gl')).default

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

    if (mapboxData.mapboxToken)
      mapboxLibrary.accessToken = mapboxData.mapboxToken

    map.current = new mapboxLibrary.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/streets-v11',
      zoom: initialZoom,
    })

    configureMapbox(map.current, mapboxLibrary)
  }, [mapContainer, mapboxLibrary, mapboxData])

  return {
    mapContainer: <MapContainer ref={mapContainer}></MapContainer>,
    mapboxMap: map.current,
    mapboxLibrary,
    mapboxToken: mapboxData?.mapboxToken || null,
  }
}

export default useMapboxMap
