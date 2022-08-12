/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { OrderDirection, MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: shareAlbumQuery
// ====================================================

export interface shareAlbumQuery_album_subAlbums_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface shareAlbumQuery_album_subAlbums_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: shareAlbumQuery_album_subAlbums_thumbnail_thumbnail | null
}

export interface shareAlbumQuery_album_subAlbums {
  __typename: 'Album'
  id: string
  title: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: shareAlbumQuery_album_subAlbums_thumbnail | null
}

export interface shareAlbumQuery_album_media_thumbnail {
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

export interface shareAlbumQuery_album_media_downloads_mediaUrl {
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
  /**
   * The file size of the resource in bytes
   */
  fileSize: number
}

export interface shareAlbumQuery_album_media_downloads {
  __typename: 'MediaDownload'
  /**
   * A description of the role of the media file
   */
  title: string
  mediaUrl: shareAlbumQuery_album_media_downloads_mediaUrl
}

export interface shareAlbumQuery_album_media_highRes {
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

export interface shareAlbumQuery_album_media_videoWeb {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface shareAlbumQuery_album_media_exif_coordinates {
  __typename: 'Coordinates'
  /**
   * GPS latitude in degrees
   */
  latitude: number
  /**
   * GPS longitude in degrees
   */
  longitude: number
}

export interface shareAlbumQuery_album_media_exif {
  __typename: 'MediaEXIF'
  id: string
  /**
   * The description of the image
   */
  description: string | null
  /**
   * The model name of the camera
   */
  camera: string | null
  /**
   * The maker of the camera
   */
  maker: string | null
  /**
   * The name of the lens
   */
  lens: string | null
  dateShot: Time | null
  /**
   * The exposure time of the image
   */
  exposure: number | null
  /**
   * The aperature stops of the image
   */
  aperture: number | null
  /**
   * The ISO setting of the image
   */
  iso: number | null
  /**
   * The focal length of the lens, when the image was taken
   */
  focalLength: number | null
  /**
   * A formatted description of the flash settings, when the image was taken
   */
  flash: number | null
  /**
   * An index describing the mode for adjusting the exposure of the image
   */
  exposureProgram: number | null
  /**
   * GPS coordinates of where the image was taken
   */
  coordinates: shareAlbumQuery_album_media_exif_coordinates | null
}

export interface shareAlbumQuery_album_media {
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
  thumbnail: shareAlbumQuery_album_media_thumbnail | null
  /**
   * A list of different versions of files for this media that can be downloaded by the user
   */
  downloads: shareAlbumQuery_album_media_downloads[]
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: shareAlbumQuery_album_media_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: shareAlbumQuery_album_media_videoWeb | null
  exif: shareAlbumQuery_album_media_exif | null
}

export interface shareAlbumQuery_album {
  __typename: 'Album'
  id: string
  title: string
  /**
   * The albums contained in this album
   */
  subAlbums: shareAlbumQuery_album_subAlbums[]
  /**
   * The media inside this album
   */
  media: shareAlbumQuery_album_media[]
}

export interface shareAlbumQuery {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: shareAlbumQuery_album
}

export interface shareAlbumQueryVariables {
  id: string
  token: string
  password?: string | null
  mediaOrderBy?: string | null
  mediaOrderDirection?: OrderDirection | null
  limit?: number | null
  offset?: number | null
}
