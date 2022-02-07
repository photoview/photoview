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
  /**
   * A bounding box of where on the image the face is present
   */
  rectangle: myFaces_myFaceGroups_imageFaces_rectangle
  /**
   * A reference to the image the face appears on
   */
  media: myFaces_myFaceGroups_imageFaces_media
}

export interface myFaces_myFaceGroups {
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
  imageFaces: myFaces_myFaceGroups_imageFaces[]
}

export interface myFaces {
  /**
   * Get a list of `FaceGroup`s for the logged in user
   */
  myFaceGroups: myFaces_myFaceGroups[]
}

export interface myFacesVariables {
  limit?: number | null
  offset?: number | null
}
