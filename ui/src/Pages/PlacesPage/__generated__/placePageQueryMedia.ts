/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: placePageQueryMedia
// ====================================================

export interface placePageQueryMedia_mediaList_thumbnail {
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

export interface placePageQueryMedia_mediaList_highRes {
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

export interface placePageQueryMedia_mediaList_videoWeb {
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

export interface placePageQueryMedia_mediaList {
  __typename: 'Media'
  id: string
  title: string
  /**
   * A short string that can be used to generate a blured version of the media, to show while the original is loading
   */
  blurhash: string | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: placePageQueryMedia_mediaList_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: placePageQueryMedia_mediaList_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: placePageQueryMedia_mediaList_videoWeb | null
  type: MediaType
}

export interface placePageQueryMedia {
  /**
   * Get a list of media by their ids, user must own the media or be admin
   */
  mediaList: placePageQueryMedia_mediaList[]
}

export interface placePageQueryMediaVariables {
  mediaIDs: string[]
}
