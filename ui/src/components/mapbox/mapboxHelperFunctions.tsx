import type mapboxgl from 'mapbox-gl'
import React from 'react'
import { createRoot } from 'react-dom/client'
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

type MarkerElement = HTMLDivElement & {
  _root?: ReturnType<typeof createRoot>
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
        if (!feature.geometry) {
          console.warn('WARN: geojson feature had no geometry', { feature })
          continue
        }

        // Type guard to ensure geometry is a Point
        if (feature.geometry.type !== 'Point') {
          console.warn('WARN: geojson feature geometry is not a Point', { feature })
          continue
        }

        const coords = feature.geometry.coordinates as [number, number]
        const props = feature.properties as MediaMarker
        if (props == null) {
          console.warn('WARN: geojson feature had no properties', {
            feature,
            geometry: feature.geometry,
            coordinates: feature.geometry.coordinates,
          })
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
          if (el) {
            marker = markers[id] = new mapboxLibrary.Marker({
              element: el,
            }).setLngLat(coords)
          } else {
            console.error('Failed to create marker element for:', id)
            continue
          }
        }
        newMarkers[id] = marker

        if (!markersOnScreen[id]) marker.addTo(map)
      }
      // for every marker we've added previously, remove those that are no longer visible
      for (const id in markersOnScreen) {
        if (!newMarkers[id]) {
          const el = markersOnScreen[id].getElement() as MarkerElement
          el._root?.unmount()
          markersOnScreen[id].remove()
        }
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
  try {
    const el = document.createElement('div') as MarkerElement
    const root = createRoot(el)
    root.render(
      <MapClusterMarker
        marker={geojsonProps}
        dispatchMarkerMedia={dispatchMarkerMedia}
      />
    )
    el._root = root
    return el
  } catch (error) {
    console.error('Failed to create cluster popup element:', error)
    // throw error or return a fallback element to make error handling more consistent
  }
}
