/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { OrderDirection, MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: albumQuery
// ====================================================

export interface albumQuery_album_subAlbums_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface albumQuery_album_subAlbums_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: albumQuery_album_subAlbums_thumbnail_thumbnail | null
}

export interface albumQuery_album_subAlbums {
  __typename: 'Album'
  id: string
  title: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: albumQuery_album_subAlbums_thumbnail | null
}

export interface albumQuery_album_media_thumbnail {
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

export interface albumQuery_album_media_highRes {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface albumQuery_album_media_videoWeb {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface albumQuery_album_media {
  __typename: 'Media'
  id: string
  type: MediaType
  /**
   * A short string that can be used to generate a blured version of the media, to show while the original is loading
   */
  blurhash: string | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: albumQuery_album_media_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: albumQuery_album_media_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: albumQuery_album_media_videoWeb | null
  favorite: boolean
}

export interface albumQuery_album {
  __typename: 'Album'
  id: string
  title: string
  /**
   * The albums contained in this album
   */
  subAlbums: albumQuery_album_subAlbums[]
  /**
   * The media inside this album
   */
  media: albumQuery_album_media[]
}

export interface albumQuery {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: albumQuery_album
}

export interface albumQueryVariables {
  id: string
  onlyFavorites?: boolean | null
  mediaOrderBy?: string | null
  orderDirection?: OrderDirection | null
  limit?: number | null
  offset?: number | null
}
