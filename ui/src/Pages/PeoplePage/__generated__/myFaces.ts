/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: myFaces
// ====================================================

export interface myFaces_myFaceGroups_imageFaces_rectangle {
  __typename: 'FaceRectangle'
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export interface myFaces_myFaceGroups_imageFaces_media_thumbnail {
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

export interface myFaces_myFaceGroups_imageFaces_media {
  __typename: 'Media'
  id: string
  title: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: myFaces_myFaceGroups_imageFaces_media_thumbnail | null
}

export interface myFaces_myFaceGroups_imageFaces {
  __typename: 'ImageFace'
  id: string
  rectangle: myFaces_myFaceGroups_imageFaces_rectangle
  media: myFaces_myFaceGroups_imageFaces_media
}

export interface myFaces_myFaceGroups {
  __typename: 'FaceGroup'
  id: string
  label: string | null
  imageFaceCount: number
  imageFaces: myFaces_myFaceGroups_imageFaces[]
}

export interface myFaces {
  myFaceGroups: myFaces_myFaceGroups[]
}

export interface myFacesVariables {
  limit?: number | null
  offset?: number | null
}
