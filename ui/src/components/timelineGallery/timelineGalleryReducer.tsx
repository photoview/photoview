import React from 'react'
import { myTimeline_myTimeline } from './__generated__/myTimeline'
import { TimelineGroup, TimelineGroupAlbum } from './TimelineGallery'
import { GalleryAction } from '../photoGallery/mediaGalleryReducer'
import { isNil } from '../../helpers/utils'

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
      const timelineGroups = convertMediaToTimelineGroups(action.timeline)

      return {
        ...state,
        activeIndex: {
          album: -1,
          date: -1,
          media: -1,
        },
        timelineGroups,
      }
    }
    case 'nextImage': {
      const { activeIndex, timelineGroups } = state

      if (
        activeIndex.album == -1 &&
        activeIndex.date == -1 &&
        activeIndex.media == -1
      ) {
        return state
      }

      const albumGroups = timelineGroups[activeIndex.date].albums
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

      if (
        activeIndex.album == -1 &&
        activeIndex.date == -1 &&
        activeIndex.media == -1
      ) {
        return state
      }

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
        const albumGroups = state.timelineGroups[activeIndex.date].albums
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
        const albumGroups = state.timelineGroups[activeIndex.date - 1].albums
        const albumMedia = albumGroups[albumGroups.length - 1].media

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
}): myTimeline_myTimeline => {
  const { date, album, media } = index
  return mediaState.timelineGroups[date].albums[album].media[media]
}

export const getActiveTimelineImage = ({
  mediaState,
}: {
  mediaState: TimelineGalleryState
}) => {
  if (
    Object.values(mediaState.activeIndex).reduce<boolean>(
      (acc, next) => next === -1 || acc,
      false
    )
  ) {
    return undefined
  }

  return getTimelineImage({ mediaState, index: mediaState.activeIndex })
}

function convertMediaToTimelineGroups(
  timelineMedia: myTimeline_myTimeline[]
): TimelineGroup[] {
  const timelineGroups: TimelineGroup[] = []
  let albums: TimelineGroupAlbum[] = []
  let nextAlbum: TimelineGroupAlbum | null = null

  const sameDay = (a: string, b: string) => {
    return (
      a.replace(/\d{2}:\d{2}:\d{2}/, '00:00:00') ==
      b.replace(/\d{2}:\d{2}:\d{2}/, '00:00:00')
    )
  }

  for (const media of timelineMedia) {
    if (nextAlbum == null) {
      nextAlbum = {
        id: media.album.id,
        title: media.album.title,
        media: [media],
      }
      continue
    }

    // if date changes
    if (!sameDay(nextAlbum.media[0].date, media.date)) {
      albums.push(nextAlbum)

      timelineGroups.push({
        date: albums[0].media[0].date.replace(/\d{2}:\d{2}:\d{2}/, '00:00:00'),
        albums: albums,
      })
      albums = []
      nextAlbum = {
        id: media.album.id,
        title: media.album.title,
        media: [media],
      }
      continue
    }

    // if album changes
    if (nextAlbum.id != media.album.id) {
      albums.push(nextAlbum)
      nextAlbum = {
        id: media.album.id,
        title: media.album.title,
        media: [media],
      }
      continue
    }

    // same album and date
    nextAlbum.media.push(media)
  }

  if (!isNil(nextAlbum)) {
    albums.push(nextAlbum)

    timelineGroups.push({
      date: albums[0].media[0].date.replace(/\d{2}:\d{2}:\d{2}/, '00:00:00'),
      albums: albums,
    })
  }

  return timelineGroups
}

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
