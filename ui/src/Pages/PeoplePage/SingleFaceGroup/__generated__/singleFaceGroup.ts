/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from './../../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: singleFaceGroup
// ====================================================

export interface singleFaceGroup_faceGroup_imageFaces_rectangle {
  __typename: 'FaceRectangle'
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export interface singleFaceGroup_faceGroup_imageFaces_media_thumbnail {
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

export interface singleFaceGroup_faceGroup_imageFaces_media_highRes {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface singleFaceGroup_faceGroup_imageFaces_media {
  __typename: 'Media'
  id: string
  type: MediaType
  title: string
  /**
   * A short string that can be used to generate a blured version of the media, to show while the original is loading
   */
  blurhash: string | null
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: singleFaceGroup_faceGroup_imageFaces_media_thumbnail | null
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: singleFaceGroup_faceGroup_imageFaces_media_highRes | null
  favorite: boolean
}

export interface singleFaceGroup_faceGroup_imageFaces {
  __typename: 'ImageFace'
  id: string
  /**
   * A bounding box of where on the image the face is present
   */
  rectangle: singleFaceGroup_faceGroup_imageFaces_rectangle
  /**
   * A reference to the image the face appears on
   */
  media: singleFaceGroup_faceGroup_imageFaces_media
}

export interface singleFaceGroup_faceGroup {
  __typename: 'FaceGroup'
  id: string
  /**
   * The name of the person
   */
  label: string | null
  imageFaces: singleFaceGroup_faceGroup_imageFaces[]
}

export interface singleFaceGroup {
  /**
   * Get a particular `FaceGroup` specified by its ID
   */
  faceGroup: singleFaceGroup_faceGroup
}

export interface singleFaceGroupVariables {
  id: string
  limit: number
  offset: number
}
