/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: myTimeline
// ====================================================

export interface myTimeline_myTimeline_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
  /**
   * Width of the image in pixels
   */
  width: number
  /**
   * Height of the image in pixels
   */
  height: number
}

export interface myTimeline_myTimeline_highRes {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
  /**
   * Width of the image in pixels
   */
  width: number
  /**
   * Height of the image in pixels
   */
  height: number
}

export interface myTimeline_myTimeline_videoWeb {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface myTimeline_myTimeline_album {
  __typename: 'Album'
  id: string
  title: string
}

export interface myTimeline_myTimeline {
  __typename: 'Media'
  id: string
  title: string
  type: MediaType
  /**
   * A short string that can be used to generate a blured version of the media, to show while the original is loading
   */
  blurhash: string | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: myTimeline_myTimeline_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: myTimeline_myTimeline_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: myTimeline_myTimeline_videoWeb | null
  favorite: boolean
  /**
   * The album that holds the media
   */
  album: myTimeline_myTimeline_album
  /**
   * The date the image was shot or the date it was imported as a fallback
   */
  date: Time
}

export interface myTimeline {
  /**
   * Get a list of media, ordered first by day, then by album if multiple media was found for the same day.
   */
  myTimeline: myTimeline_myTimeline[]
}

export interface myTimelineVariables {
  onlyFavorites?: boolean | null
  limit?: number | null
  offset?: number | null
  fromDate?: Time | null
}
