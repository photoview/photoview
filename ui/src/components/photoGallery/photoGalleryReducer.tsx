import React from 'react'
import { UpdateSidebarFn } from '../sidebar/Sidebar'
import { PhotoGalleryProps_Media } from './PhotoGallery'
import MediaSidebar from '../sidebar/MediaSidebar'

export type PhotoGalleryState = {
  presenting: boolean
  activeIndex: number
  media: PhotoGalleryProps_Media[]
}

export type PhotoGalleryAction =
  | { type: 'nextImage' }
  | { type: 'previousImage' }
  | { type: 'setPresenting'; presenting: boolean }
  | { type: 'selectImage'; index: number }

export function photoGalleryReducer(
  state: PhotoGalleryState,
  action: PhotoGalleryAction
): PhotoGalleryState {
  switch (action.type) {
    case 'nextImage':
      return {
        ...state,
        activeIndex: (state.activeIndex + 1) % state.media.length,
      }
    case 'previousImage':
      if (state.activeIndex <= 0) {
        return {
          ...state,
          activeIndex: state.media.length - 1,
        }
      } else {
        return {
          ...state,
          activeIndex: state.activeIndex - 1,
        }
      }
    case 'setPresenting':
      return {
        ...state,
        presenting: action.presenting,
      }
    case 'selectImage':
      return {
        ...state,
        activeIndex: Math.max(
          0,
          Math.min(state.media.length - 1, action.index)
        ),
      }
  }
}

export const selectImageAction = ({
  index,
  mediaState,
  dispatchMedia,
  updateSidebar,
}: {
  index: number
  mediaState: PhotoGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
  updateSidebar: UpdateSidebarFn
}) => {
  updateSidebar(
    <MediaSidebar media={mediaState.media[mediaState.activeIndex]} />
  )
  dispatchMedia({
    type: 'selectImage',
    index,
  })
}
