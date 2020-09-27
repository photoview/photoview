import React from 'react'
import ReactDOM from 'react-dom'
import MapClusterMarker from './MapClusterMarker'

let markers = {}
let markersOnScreen = {}

export const makeUpdateMarkers = ({
  map,
  mapboxLibrary,
  setPresentMarker,
}) => () => {
  let newMarkers = {}
  const features = map.querySourceFeatures('media')

  // for every media on the screen, create an HTML marker for it (if we didn't yet),
  // and add it to the map if it's not there already
  for (let i = 0; i < features.length; i++) {
    const coords = features[i].geometry.coordinates
    const props = features[i].properties
    const id = props.cluster ? props.cluster_id : props.media_id

    let marker = markers[id]
    if (!marker) {
      let el = createClusterPopupElement(props, setPresentMarker)
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

function createClusterPopupElement(geojsonProps, setPresentMarker) {
  const el = document.createElement('div')
  ReactDOM.render(
    <MapClusterMarker {...geojsonProps} setPresentMarker={setPresentMarker} />,
    el
  )
  return el
}
