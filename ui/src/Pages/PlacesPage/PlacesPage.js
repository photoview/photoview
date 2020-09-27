import React, { useEffect, useRef, useState } from 'react'
import ReactDOM from 'react-dom'
import PropTypes from 'prop-types'
import { useQuery } from 'react-apollo'
import gql from 'graphql-tag'
import styled from 'styled-components'

import 'mapbox-gl/dist/mapbox-gl.css'

import Layout from '../../Layout'

import MapClusterMarker from './MapClusterMarker'

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

      map.current.on('move', updateMarkers)
      map.current.on('moveend', updateMarkers)
      updateMarkers()

      var markers = {}
      var markersOnScreen = {}

      function updateMarkers() {
        var newMarkers = {}
        var features = map.current.querySourceFeatures('media')

        // for every media on the screen, create an HTML marker for it (if we didn't yet),
        // and add it to the map if it's not there already
        for (var i = 0; i < features.length; i++) {
          var coords = features[i].geometry.coordinates
          var props = features[i].properties
          var id = props.cluster ? props.cluster_id : props.media_id

          var marker = markers[id]
          if (!marker) {
            var el = createClusterPopupElement(props)
            marker = markers[id] = new mapboxLibrary.Marker({
              element: el,
            }).setLngLat(coords)
          }
          newMarkers[id] = marker

          if (!markersOnScreen[id]) marker.addTo(map.current)
        }
        // for every marker we've added previously, remove those that are no longer visible
        for (id in markersOnScreen) {
          if (!newMarkers[id]) markersOnScreen[id].remove()
        }
        markersOnScreen = newMarkers
      }

      function createClusterPopupElement(props) {
        const el = document.createElement('div')
        ReactDOM.render(<MapClusterMarker {...props} />, el)
        return el
      }

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
