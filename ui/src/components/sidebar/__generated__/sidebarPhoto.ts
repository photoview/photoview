/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: sidebarPhoto
// ====================================================

export interface sidebarPhoto_media_highRes {
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

export interface sidebarPhoto_media_thumbnail {
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

export interface sidebarPhoto_media_videoWeb {
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

export interface sidebarPhoto_media_videoMetadata {
  __typename: 'VideoMetadata'
  id: string
  width: number
  height: number
  duration: number
  codec: string | null
  framerate: number | null
  bitrate: string | null
  colorProfile: string | null
  audio: string | null
}

export interface sidebarPhoto_media_exif {
  __typename: 'MediaEXIF'
  id: string
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
  dateShot: any | null
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
}

export interface sidebarPhoto_media_faces_rectangle {
  __typename: 'FaceRectangle'
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export interface sidebarPhoto_media_faces_faceGroup {
  __typename: 'FaceGroup'
  id: string
}

export interface sidebarPhoto_media_faces {
  __typename: 'ImageFace'
  id: string
  rectangle: sidebarPhoto_media_faces_rectangle
  faceGroup: sidebarPhoto_media_faces_faceGroup
}

export interface sidebarPhoto_media {
  __typename: 'Media'
  id: string
  title: string
  type: MediaType
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: sidebarPhoto_media_highRes | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: sidebarPhoto_media_thumbnail | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: sidebarPhoto_media_videoWeb | null
  videoMetadata: sidebarPhoto_media_videoMetadata | null
  exif: sidebarPhoto_media_exif | null
  faces: sidebarPhoto_media_faces[]
}

export interface sidebarPhoto {
  /**
   * Get media by id, user must own the media or be admin.
   * If valid tokenCredentials are provided, the media may be retrived without further authentication
   */
  media: sidebarPhoto_media
}

export interface sidebarPhotoVariables {
  id: string
}
