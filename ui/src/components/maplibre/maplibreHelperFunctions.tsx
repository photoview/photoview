import type maplibregl from 'maplibre-gl'
import type geojson from 'geojson'
import React from 'react'
import { createRoot, Root } from 'react-dom/client'
import MapClusterMarker from '../../Pages/PlacesPage/MapClusterMarker'
import { MediaMarker } from '../../Pages/PlacesPage/MapPresentMarker'
import { PlacesAction } from '../../Pages/PlacesPage/placesReducer'

type MarkerEntry = {
  marker: maplibregl.Marker
  root: Root
}

type registerMediaMarkersArgs = {
  map: maplibregl.Map
  maplibreLibrary: typeof maplibregl
  dispatchMarkerMedia: React.Dispatch<PlacesAction>
}

/**
 * Add appropriate event handlers to the map, to render and update media markers
 * Expects the provided map to contain geojson source of media
 */
export const registerMediaMarkers = ({
  map,
  maplibreLibrary,
  dispatchMarkerMedia,
}: registerMediaMarkersArgs): (() => void) => {
  const markers: { [key: string]: MarkerEntry } = {}
  let markersOnScreen: { [key: string]: MarkerEntry } = {}

  const updateMarkers = () => {
    const newMarkers: { [key: string]: MarkerEntry } = {}
    const features = map.querySourceFeatures('media')

    // for every media on the screen, create an HTML marker for it (if we didn't yet),
    // and add it to the map if it's not there already
    for (const feature of features) {
      const point = feature.geometry as geojson.Point
      const coords = point.coordinates as [number, number]
      const props = feature.properties as MediaMarker
      if (props == null) {
        console.warn('WARN: geojson feature had no properties', feature)
        continue
      }

      const id = props.cluster
        ? `cluster_${props.cluster_id}`
        : `media_${props.media_id}`

      let entry = markers[id]
      if (!entry) {
        const { el, root } = createClusterPopupElement(props, {
          dispatchMarkerMedia,
        })
        const marker = new maplibreLibrary.Marker({
          element: el,
        }).setLngLat(coords)
        entry = markers[id] = { marker, root }
      }
      newMarkers[id] = entry

      if (!markersOnScreen[id]) entry.marker.addTo(map)
    }
    // for every marker we've added previously, remove those that are no longer visible
    for (const id in markersOnScreen) {
      if (!newMarkers[id]) {
        markersOnScreen[id].marker.remove()
        markersOnScreen[id].root.unmount()
        delete markers[id]
      }
    }
    markersOnScreen = newMarkers
  }

  map.on('move', updateMarkers)
  map.on('moveend', updateMarkers)
  map.on('sourcedata', updateMarkers)
  updateMarkers()

  return () => {
    map.off('move', updateMarkers)
    map.off('moveend', updateMarkers)
    map.off('sourcedata', updateMarkers)
    for (const id in markers) {
      markers[id].marker.remove()
      markers[id].root.unmount()
    }
  }
}

function createClusterPopupElement(
  geojsonProps: MediaMarker,
  {
    dispatchMarkerMedia,
  }: {
    dispatchMarkerMedia: React.Dispatch<PlacesAction>
  }
) {
  const el = document.createElement('div')
  const root = createRoot(el)
  root.render(
    <MapClusterMarker
      marker={geojsonProps}
      dispatchMarkerMedia={dispatchMarkerMedia}
    />
  )
  return { el, root }
}
