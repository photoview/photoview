/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: SharePageToken
// ====================================================

export interface SharePageToken_shareToken_album {
  __typename: 'Album'
  id: string
}

export interface SharePageToken_shareToken_media_thumbnail {
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

export interface SharePageToken_shareToken_media_downloads_mediaUrl {
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

export interface SharePageToken_shareToken_media_downloads {
  __typename: 'MediaDownload'
  /**
   * A description of the role of the media file
   */
  title: string
  mediaUrl: SharePageToken_shareToken_media_downloads_mediaUrl
}

export interface SharePageToken_shareToken_media_highRes {
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

export interface SharePageToken_shareToken_media_videoWeb {
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

export interface SharePageToken_shareToken_media_exif_coordinates {
  __typename: 'Coordinates'
  /**
   * GPS longitude in degrees
   */
  longitude: number
  /**
   * GPS latitude in degrees
   */
  latitude: number
}

export interface SharePageToken_shareToken_media_exif {
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
  coordinates: SharePageToken_shareToken_media_exif_coordinates | null
}

export interface SharePageToken_shareToken_media {
  __typename: 'Media'
  id: string
  title: string
  type: MediaType
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: SharePageToken_shareToken_media_thumbnail | null
  /**
   * A list of different versions of files for this media that can be downloaded by the user
   */
  downloads: SharePageToken_shareToken_media_downloads[]
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: SharePageToken_shareToken_media_highRes | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: SharePageToken_shareToken_media_videoWeb | null
  exif: SharePageToken_shareToken_media_exif | null
}

export interface SharePageToken_shareToken {
  __typename: 'ShareToken'
  token: string
  /**
   * The album this token shares
   */
  album: SharePageToken_shareToken_album | null
  /**
   * The media this token shares
   */
  media: SharePageToken_shareToken_media | null
}

export interface SharePageToken {
  /**
   * Fetch a share token containing an `Album` or `Media`
   */
  shareToken: SharePageToken_shareToken
}

export interface SharePageTokenVariables {
  token: string
  password?: string | null
}
