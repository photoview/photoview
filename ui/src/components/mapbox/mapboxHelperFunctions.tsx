import type mapboxgl from 'mapbox-gl'
import type geojson from 'geojson'
import React from 'react'
import ReactDOM from 'react-dom'
import MapClusterMarker from '../../Pages/PlacesPage/MapClusterMarker'
import { MediaMarker } from '../../Pages/PlacesPage/MapPresentMarker'
import { PlacesAction } from '../../Pages/PlacesPage/placesReducer'

const markers: { [key: string]: mapboxgl.Marker } = {}
let markersOnScreen: typeof markers = {}

type registerMediaMarkersArgs = {
  map: mapboxgl.Map
  mapboxLibrary: typeof mapboxgl
  dispatchMarkerMedia: React.Dispatch<PlacesAction>
}

/**
 * Add appropriate event handlers to the map, to render and update media markers
 * Expects the provided mapbox map to contain geojson source of media
 */
export const registerMediaMarkers = (args: registerMediaMarkersArgs) => {
  const updateMarkers = makeUpdateMarkers(args)

  args.map.on('move', updateMarkers)
  args.map.on('moveend', updateMarkers)
  args.map.on('sourcedata', updateMarkers)
  updateMarkers()
}

/**
 * Make a function that can be passed to Mapbox to tell it how to render and update the image markers
 */
const makeUpdateMarkers =
  ({ map, mapboxLibrary, dispatchMarkerMedia }: registerMediaMarkersArgs) =>
  () => {
    const newMarkers: typeof markers = {}
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

      let marker = markers[id]
      if (!marker) {
        const el = createClusterPopupElement(props, {
          dispatchMarkerMedia,
        })
        marker = markers[id] = new mapboxLibrary.Marker({
          element: el,
        }).setLngLat(coords)
      }
      newMarkers[id] = marker

      if (!markersOnScreen[id]) marker.addTo(map)
    }
    // for every marker we've added previously, remove those that are no longer visible
    for (const id in markersOnScreen) {
      if (!newMarkers[id]) markersOnScreen[id].remove()
    }
    markersOnScreen = newMarkers
  }

function createClusterPopupElement(
  geojsonProps: MediaMarker,
  {
    dispatchMarkerMedia,
  }: {
    dispatchMarkerMedia: React.Dispatch<PlacesAction>
  }
) {
  // setPresentMarker: React.Dispatch<React.SetStateAction<PresentMarker | null>>
  const el = document.createElement('div')
  ReactDOM.render(
    <MapClusterMarker
      marker={geojsonProps}
      dispatchMarkerMedia={dispatchMarkerMedia}
    />,
    el
  )
  return el
}
