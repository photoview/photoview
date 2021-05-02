import { PresentMarker } from './PlacesPage'
import {
  PhotoGalleryState,
  PhotoGalleryAction,
  photoGalleryReducer,
} from './../../components/photoGallery/photoGalleryReducer'

export interface PlacesState extends PhotoGalleryState {
  presentMarker?: PresentMarker
}

export type PlacesAction =
  | PhotoGalleryAction
  | { type: 'replacePresentMarker'; marker?: PresentMarker }

export function placesReducer(
  state: PlacesState,
  action: PlacesAction
): PlacesState {
  switch (action.type) {
    case 'replacePresentMarker':
      if (
        state.presentMarker &&
        action.marker &&
        state.presentMarker.cluster === action.marker.cluster &&
        state.presentMarker.id === action.marker.id
      ) {
        return {
          ...state,
          presenting: true,
        }
      } else {
        return {
          ...state,
          presentMarker: action.marker,
        }
      }
    default:
      return photoGalleryReducer(state, action)
  }
}

// export const presentMarkerClicked = ({
//   dispatchMedia,
//   mediaState,
//   marker,
// }: {
//   dispatchMedia: React.Dispatch<PlacesAction>
//   mediaState: PlacesState
//   marker: PresentMarker
// }) => {
//   console.log(
//     'present marker clicked',
//     mediaState,
//     marker,
//     !!mediaState.presentMarker,
//     mediaState.presentMarker?.cluster === marker.cluster,
//     mediaState.presentMarker?.id === marker.id
//   )
//   if (
//     mediaState.presentMarker &&
//     mediaState.presentMarker.cluster === marker.cluster &&
//     mediaState.presentMarker.id === marker.id
//   ) {
//     console.log('EQUAL OPEN PRESENT MODE')
//     history.pushState(
//       { presenting: true, activeIndex: mediaState.activeIndex },
//       ''
//     )
//     dispatchMedia({
//       type: 'openPresentMode',
//       activeIndex: mediaState.activeIndex,
//     })
//   } else {
//     dispatchMedia({
//       type: 'replacePresentMarker',
//       marker,
//     })
//   }
// }
