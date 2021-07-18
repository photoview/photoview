/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: myTimeline
// ====================================================

export interface myTimeline_myTimeline_album {
  __typename: 'Album'
  id: string
  title: string
}

export interface myTimeline_myTimeline_media_thumbnail {
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

export interface myTimeline_myTimeline_media_highRes {
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

export interface myTimeline_myTimeline_media_videoWeb {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface myTimeline_myTimeline_media {
  __typename: 'Media'
  id: string
  title: string
  type: MediaType
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: myTimeline_myTimeline_media_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: myTimeline_myTimeline_media_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: myTimeline_myTimeline_media_videoWeb | null
  favorite: boolean
}

export interface myTimeline_myTimeline {
  __typename: 'TimelineGroup'
  album: myTimeline_myTimeline_album
  media: myTimeline_myTimeline_media[]
  mediaTotal: number
  date: any
}

export interface myTimeline {
  myTimeline: myTimeline_myTimeline[]
}

export interface myTimelineVariables {
  onlyFavorites?: boolean | null
  limit?: number | null
  offset?: number | null
}
