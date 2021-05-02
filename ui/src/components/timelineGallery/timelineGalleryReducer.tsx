import React from 'react'
import {
  myTimeline_myTimeline,
  myTimeline_myTimeline_media,
} from './__generated__/myTimeline'
import { TimelineGroup } from './TimelineGallery'
import { GalleryAction } from '../photoGallery/photoGalleryReducer'

export interface TimelineMediaIndex {
  date: number
  album: number
  media: number
}

export interface TimelineGalleryState {
  presenting: boolean
  timelineGroups: TimelineGroup[]
  activeIndex: TimelineMediaIndex
}

export type TimelineGalleryAction =
  | GalleryAction
  | { type: 'replaceTimelineGroups'; timeline: myTimeline_myTimeline[] }
  | { type: 'selectImage'; index: TimelineMediaIndex }
  | { type: 'openPresentMode'; activeIndex: TimelineMediaIndex }

export function timelineGalleryReducer(
  state: TimelineGalleryState,
  action: TimelineGalleryAction
): TimelineGalleryState {
  switch (action.type) {
    case 'replaceTimelineGroups': {
      const dateGroupedAlbums = action.timeline.reduce((acc, val) => {
        if (acc.length == 0 || acc[acc.length - 1].date != val.date) {
          acc.push({
            date: val.date,
            groups: [val],
          })
        } else {
          acc[acc.length - 1].groups.push(val)
        }

        return acc
      }, [] as TimelineGroup[])

      return {
        ...state,
        activeIndex: {
          album: -1,
          date: -1,
          media: -1,
        },
        timelineGroups: dateGroupedAlbums,
      }
    }
    case 'nextImage': {
      const { activeIndex, timelineGroups } = state

      const albumGroups = timelineGroups[activeIndex.date].groups
      const albumMedia = albumGroups[activeIndex.album].media

      if (activeIndex.media < albumMedia.length - 1) {
        return {
          ...state,
          activeIndex: {
            ...state.activeIndex,
            media: activeIndex.media + 1,
          },
        }
      }

      if (activeIndex.album < albumGroups.length - 1) {
        return {
          ...state,
          activeIndex: {
            ...state.activeIndex,
            album: activeIndex.album + 1,
            media: 0,
          },
        }
      }

      if (activeIndex.date < timelineGroups.length - 1) {
        return {
          ...state,
          activeIndex: {
            date: activeIndex.date + 1,
            album: 0,
            media: 0,
          },
        }
      }

      // reached the end
      return state
    }
    case 'previousImage': {
      const { activeIndex } = state

      if (activeIndex.media > 0) {
        return {
          ...state,
          activeIndex: {
            ...activeIndex,
            media: activeIndex.media - 1,
          },
        }
      }

      if (activeIndex.album > 0) {
        const albumGroups = state.timelineGroups[activeIndex.date].groups
        const albumMedia = albumGroups[activeIndex.album - 1].media

        return {
          ...state,
          activeIndex: {
            ...activeIndex,
            album: activeIndex.album - 1,
            media: albumMedia.length - 1,
          },
        }
      }

      if (activeIndex.date > 0) {
        const albumGroups = state.timelineGroups[activeIndex.date - 1].groups
        const albumMedia = albumGroups[activeIndex.album - 1].media

        return {
          ...state,
          activeIndex: {
            date: activeIndex.date - 1,
            album: albumGroups.length - 1,
            media: albumMedia.length - 1,
          },
        }
      }

      // reached the start
      return state
    }
    case 'selectImage': {
      return {
        ...state,
        activeIndex: action.index,
      }
    }
    case 'openPresentMode':
      return {
        ...state,
        presenting: true,
        activeIndex: action.activeIndex,
      }
    case 'closePresentMode': {
      return {
        ...state,
        presenting: false,
      }
    }
  }
}

export const getTimelineImage = ({
  mediaState,
  index,
}: {
  mediaState: TimelineGalleryState
  index: TimelineMediaIndex
}): myTimeline_myTimeline_media => {
  const { date, album, media } = index
  return mediaState.timelineGroups[date].groups[album].media[media]
}

export const getActiveTimelineImage = ({
  mediaState,
}: {
  mediaState: TimelineGalleryState
}) => {
  if (
    Object.values(mediaState.activeIndex).reduce(
      (acc, next) => next == -1 || acc,
      false
    )
  ) {
    return undefined
  }

  return getTimelineImage({ mediaState, index: mediaState.activeIndex })
}

// export const selectTimelineImageAction = ({
//   index,
//   mediaState,
//   dispatchMedia,
//   updateSidebar,
// }: {
//   index: TimelineMediaIndex
//   mediaState: TimelineGalleryState
//   dispatchMedia: React.Dispatch<TimelineGalleryAction>
//   updateSidebar: UpdateSidebarFn
// }) => {
//   updateSidebar(
//     <MediaSidebar media={getTimelineImage({ mediaState, index })} />
//   )
//   dispatchMedia({
//     type: 'selectImage',
//     index,
//   })
// }

export const openTimelinePresentMode = ({
  dispatchMedia,
  activeIndex,
}: {
  dispatchMedia: React.Dispatch<TimelineGalleryAction>
  activeIndex: TimelineMediaIndex
}) => {
  dispatchMedia({
    type: 'openPresentMode',
    activeIndex,
  })

  history.pushState({ presenting: true, activeIndex: activeIndex }, '')
}
