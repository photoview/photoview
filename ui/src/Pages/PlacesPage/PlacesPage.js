import React, { useEffect, useRef, useState } from 'react'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import styled from 'styled-components'

import 'mapbox-gl/dist/mapbox-gl.css'

import Layout from '../../Layout'
import imagePopup from './image-popup.png'

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
      // center: [this.state.lng, this.state.lat],
      // zoom: this.state.zoom
    })

    map.current.on('load', () => {
      console.log(mapboxData.myMediaGeoJson)
      map.current.addSource('media', {
        type: 'geojson',
        data: mapboxData.myMediaGeoJson,
        cluster: true,
        clusterMaxZoom: 14, // Max zoom to cluster points on
        clusterRadius: 50,
        clusterProperties: {
          thumbnail_url: [
            'coalesce',
            ['get', 'url', ['get', 'thumbnail']],
            false,
          ],
        },
      })

      map.current.loadImage(imagePopup, (error, image) => {
        console.log(error, image)
        map.current.addImage('media-popup-bg', image)

        map.current.addLayer({
          id: 'media-cluster-popup',
          type: 'symbol',
          source: 'media',
          filter: ['has', 'point_count'],
          layout: {
            'icon-image': 'media-popup-bg',
            'icon-size': 0.5,
            'icon-allow-overlap': true,
          },
        })

        map.current.addLayer({
          id: 'media-cluster-count-bg',
          type: 'circle',
          source: 'media',
          filter: ['has', 'point_count'],
          paint: {
            'circle-color': '#11b4da',
            'circle-radius': 11,
            'circle-translate': [22, -24],
          },
        })

        map.current.addLayer({
          id: 'media-cluster-count',
          type: 'symbol',
          source: 'media',
          filter: ['has', 'point_count'],
          layout: {
            'text-field': '{point_count_abbreviated}',
            'text-size': 12,
            'text-allow-overlap': true,
            'text-offset': [22 / 12, -24 / 12],
          },
          paint: {
            'text-color': '#ffffff',
          },
        })
      })

      // map.current.addLayer({
      //   id: 'media-points',
      //   type: 'circle',
      //   source: 'media',
      //   filter: ['!', ['has', 'point_count']],
      //   paint: {
      //     'circle-color': '#11b4da',
      //     'circle-radius': 4,
      //     'circle-stroke-width': 1,
      //     'circle-stroke-color': '#fff',
      //   },
      // })

      console.log(map.current)
    })
  }, [mapContainer, mapboxLibrary, mapboxData])

  return (
    <Layout>
      <MapWrapper>
        <MapContainer ref={mapContainer}></MapContainer>
      </MapWrapper>
    </Layout>
  )
}

export default MapPage
