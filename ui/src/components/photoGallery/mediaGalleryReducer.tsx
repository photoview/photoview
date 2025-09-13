import React, { useEffect } from 'react'
import { MediaGalleryFields } from './__generated__/MediaGalleryFields'
import { MediaType } from '../../__generated__/globalTypes'


export interface MediaGalleryState {
  videoMediaIndices: number[]
  photoMediaIndices: number[]
  presenting: boolean
  activeIndex: number
  media: MediaGalleryFields[]
}

export type GalleryAction =
  | { type: 'nextImage' }
  | { type: 'previousImage' }
  | { type: 'closePresentMode' }
  | { type: 'nextSlidePhoto' }
  | { type: 'nextSlideVideo' }

export type PhotoGalleryAction =
  | GalleryAction
  | { type: 'openPresentMode'; activeIndex: number }
  | { type: 'selectImage'; index: number }
  | { type: 'replaceMedia'; media: MediaGalleryFields[] }

function getNextSlideState(state: MediaGalleryState, type: MediaType): MediaGalleryState {
  const indices = type === MediaType.Photo ? state.photoMediaIndices : state.videoMediaIndices
  let nextIndex = indices.find(i => i > state.activeIndex) ?? indices.at(0)

  if (nextIndex === undefined)
    return state
  
  return {
    ...state,
    activeIndex: nextIndex
  }
}

export function mediaGalleryReducer(
  state: MediaGalleryState,
  action: PhotoGalleryAction
): MediaGalleryState {
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
    case 'openPresentMode':
      return {
        ...state,
        presenting: true,
        activeIndex: action.activeIndex,
      }
    case 'closePresentMode':
      return {
        ...state,
        presenting: false,
      }
    case 'selectImage':
      return {
        ...state,
        activeIndex: Math.max(
          0,
          Math.min(state.media.length - 1, action.index)
        ),
      }
    case 'replaceMedia':
      return {
        ...state,
        media: action.media,
        activeIndex: -1,
        presenting: false,
      }
    case 'nextSlidePhoto':
     // Do nothing if all media are videos
     if (state.photoMediaIndices.length === 0)
      return {...state}
     
     return getNextSlideState(state, MediaType.Photo)
      
    case 'nextSlideVideo':
      // Do nothing if there are no videos
      if (state.videoMediaIndices.length === 0)
        return {...state}
      
      return getNextSlideState(state, MediaType.Video)
  }
}

export interface MediaGalleryPopStateEvent extends PopStateEvent {
  state: MediaGalleryState
}

export const urlPresentModeSetupHook = ({
  dispatchMedia,
  openPresentMode,
}: {
  dispatchMedia: React.Dispatch<GalleryAction>
  openPresentMode: (event: MediaGalleryPopStateEvent) => void
}) => {
  useEffect(() => {
    const urlChangeListener = (event: MediaGalleryPopStateEvent) => {
      if (event.state.presenting === true) {
        openPresentMode(event)
      } else {
        dispatchMedia({ type: 'closePresentMode' })
      }
    }

    window.addEventListener('popstate', urlChangeListener)

    history.replaceState({ presenting: false }, '')

    return () => {
      window.removeEventListener('popstate', urlChangeListener)
    }
  }, [])
}

export const openPresentModeAction = ({
  dispatchMedia,
  activeIndex,
}: {
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
  activeIndex: number
}) => {
  dispatchMedia({
    type: 'openPresentMode',
    activeIndex: activeIndex,
  })

  history.pushState({ presenting: true, activeIndex }, '')
}

export const closePresentModeAction = ({
  dispatchMedia,
}: {
  dispatchMedia: React.Dispatch<GalleryAction>
}) => {
  dispatchMedia({
    type: 'closePresentMode',
  })

  history.back()
}
