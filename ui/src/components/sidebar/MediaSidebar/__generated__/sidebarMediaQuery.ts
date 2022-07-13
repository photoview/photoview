/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: sidebarMediaQuery
// ====================================================

export interface sidebarMediaQuery_media_highRes {
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

export interface sidebarMediaQuery_media_thumbnail {
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

export interface sidebarMediaQuery_media_videoWeb {
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

export interface sidebarMediaQuery_media_videoMetadata {
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

export interface sidebarMediaQuery_media_exif_coordinates {
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

export interface sidebarMediaQuery_media_exif {
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
  coordinates: sidebarMediaQuery_media_exif_coordinates | null
}

export interface sidebarMediaQuery_media_album_path {
  __typename: 'Album'
  id: string
  title: string
}

export interface sidebarMediaQuery_media_album {
  __typename: 'Album'
  id: string
  title: string
  /**
   * A breadcrumb list of all parent albums down to this one
   */
  path: sidebarMediaQuery_media_album_path[]
}

export interface sidebarMediaQuery_media_faces_rectangle {
  __typename: 'FaceRectangle'
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export interface sidebarMediaQuery_media_faces_faceGroup {
  __typename: 'FaceGroup'
  id: string
  /**
   * The name of the person
   */
  label: string | null
  /**
   * The total number of images in this collection
   */
  imageFaceCount: number
}

export interface sidebarMediaQuery_media_faces_media_thumbnail {
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

export interface sidebarMediaQuery_media_faces_media {
  __typename: 'Media'
  id: string
  title: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: sidebarMediaQuery_media_faces_media_thumbnail | null
}

export interface sidebarMediaQuery_media_faces {
  __typename: 'ImageFace'
  id: string
  /**
   * A bounding box of where on the image the face is present
   */
  rectangle: sidebarMediaQuery_media_faces_rectangle
  /**
   * The `FaceGroup` that contains this `ImageFace`
   */
  faceGroup: sidebarMediaQuery_media_faces_faceGroup
  /**
   * A reference to the image the face appears on
   */
  media: sidebarMediaQuery_media_faces_media
}

export interface sidebarMediaQuery_media {
  __typename: 'Media'
  id: string
  title: string
  type: MediaType
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: sidebarMediaQuery_media_highRes | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: sidebarMediaQuery_media_thumbnail | null
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: sidebarMediaQuery_media_videoWeb | null
  videoMetadata: sidebarMediaQuery_media_videoMetadata | null
  exif: sidebarMediaQuery_media_exif | null
  /**
   * The album that holds the media
   */
  album: sidebarMediaQuery_media_album
  /**
   * A list of faces present on the image
   */
  faces: sidebarMediaQuery_media_faces[]
}

export interface sidebarMediaQuery {
  /**
   * Get media by id, user must own the media or be admin.
   * If valid tokenCredentials are provided, the media may be retrived without further authentication
   */
  media: sidebarMediaQuery_media
}

export interface sidebarMediaQueryVariables {
  id: string
}
