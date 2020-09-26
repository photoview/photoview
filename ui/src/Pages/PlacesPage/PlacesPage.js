import React, { useEffect, useRef, useState } from 'react'
import Layout from '../../Layout'

import 'mapbox-gl/dist/mapbox-gl.css'
import styled from 'styled-components'

const MapWrapper = styled.div`
  width: 100%;
  height: calc(100% - 24px);
`

const MapContainer = styled.div`
  width: 100%;
  height: 100%;
`

const MapPage = () => {
  const [mapboxLibrary, setMapboxLibrary] = useState(null)
  const mapContainer = useRef()
  const map = useRef()

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
      map.current != null
    )
      return

    map.current = new mapboxLibrary.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/streets-v11',
      // center: [this.state.lng, this.state.lat],
      // zoom: this.state.zoom
    })
  }, [mapContainer, mapboxLibrary])

  return (
    <Layout>
      <MapWrapper>
        <MapContainer ref={mapContainer}></MapContainer>
      </MapWrapper>
    </Layout>
  )
}

export default MapPage
