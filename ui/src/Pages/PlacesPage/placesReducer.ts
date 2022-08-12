import { PresentMarker } from './PlacesPage'
import {
  MediaGalleryState,
  PhotoGalleryAction,
  mediaGalleryReducer,
} from '../../components/photoGallery/mediaGalleryReducer'

export interface PlacesState extends MediaGalleryState {
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
      return mediaGalleryReducer(state, action)
  }
}
