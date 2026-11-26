import React, { useEffect } from 'react'
import { MediaGalleryFields } from './__generated__/MediaGalleryFields'

export interface MediaGalleryState {
  presenting: boolean
  activeIndex: number
  media: MediaGalleryFields[]
}

export type GalleryAction =
  | { type: 'nextImage' }
  | { type: 'previousImage' }
  | { type: 'closePresentMode' }

export type PhotoGalleryAction =
  | GalleryAction
  | { type: 'openPresentMode'; activeIndex: number }
  | { type: 'selectImage'; index: number }
  | { type: 'replaceMedia'; media: MediaGalleryFields[] }

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
  }
}

export interface MediaGalleryPopStateEvent extends PopStateEvent {
  state: MediaGalleryState & { groupId?: string }
}

// groupId tells apart several independent galleries on one page (the search
// results page renders one per album). Every mounted gallery listens on the
// same global popstate event, so without it going back would open present
// mode in all of them at once. Pages with a single gallery can omit it.
export const urlPresentModeSetupHook = ({
  dispatchMedia,
  openPresentMode,
  groupId,
}: {
  dispatchMedia: React.Dispatch<GalleryAction>
  openPresentMode: (event: MediaGalleryPopStateEvent) => void
  groupId?: string
}) => {
  useEffect(() => {
    const urlChangeListener = (event: MediaGalleryPopStateEvent) => {
      // Only this group's own entry opens it, but any other state closes it,
      // so a gallery left presenting doesn't stay open when navigating back
      // past it.
      if (event.state?.presenting === true && event.state.groupId === groupId) {
        openPresentMode(event)
      } else {
        dispatchMedia({ type: 'closePresentMode' })
      }
    }

    window.addEventListener('popstate', urlChangeListener)

    history.replaceState({ presenting: false, groupId }, '')

    return () => {
      window.removeEventListener('popstate', urlChangeListener)
    }
  }, [])
}

export const openPresentModeAction = ({
  dispatchMedia,
  activeIndex,
  groupId,
}: {
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
  activeIndex: number
  groupId?: string
}) => {
  dispatchMedia({
    type: 'openPresentMode',
    activeIndex: activeIndex,
  })

  history.pushState({ presenting: true, activeIndex, groupId }, '')
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
