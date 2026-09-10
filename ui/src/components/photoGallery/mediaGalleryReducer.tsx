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

// groupId distinguishes multiple independent galleries living on the same
// page (e.g. one per album on the search results page) so that browser
// back/forward only affects the gallery that pushed that history entry.
// Pages with a single gallery (the common case) can omit it.
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
      // Multiple groups can be mounted at once (e.g. one per album on the
      // search results page), each with its own base history entry that
      // the others may since have replaced. Only open when this group's id
      // matches, but always close on any other state so a group left
      // presenting doesn't get stuck open when navigating back past it.
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
