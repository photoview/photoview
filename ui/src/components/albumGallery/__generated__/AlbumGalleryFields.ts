/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL fragment: AlbumGalleryFields
// ====================================================

export interface AlbumGalleryFields_subAlbums_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface AlbumGalleryFields_subAlbums_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: AlbumGalleryFields_subAlbums_thumbnail_thumbnail | null
}

export interface AlbumGalleryFields_subAlbums {
  __typename: 'Album'
  id: string
  title: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: AlbumGalleryFields_subAlbums_thumbnail | null
}

export interface AlbumGalleryFields_media_thumbnail {
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

export interface AlbumGalleryFields_media_highRes {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface AlbumGalleryFields_media_videoWeb {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface AlbumGalleryFields_media {
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
  thumbnail: AlbumGalleryFields_media_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: AlbumGalleryFields_media_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: AlbumGalleryFields_media_videoWeb | null
  favorite: boolean
}

export interface AlbumGalleryFields {
  __typename: 'Album'
  id: string
  title: string
  /**
   * The albums contained in this album
   */
  subAlbums: AlbumGalleryFields_subAlbums[]
  /**
   * The media inside this album
   */
  media: AlbumGalleryFields_media[]
}
