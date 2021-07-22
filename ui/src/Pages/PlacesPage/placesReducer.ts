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
