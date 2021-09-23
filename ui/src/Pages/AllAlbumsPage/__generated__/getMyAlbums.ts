/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: getMyAlbums
// ====================================================

export interface getMyAlbums_myAlbums_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface getMyAlbums_myAlbums_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: getMyAlbums_myAlbums_thumbnail_thumbnail | null
}

export interface getMyAlbums_myAlbums {
  __typename: 'Album'
  id: string
  title: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: getMyAlbums_myAlbums_thumbnail | null
}

export interface getMyAlbums {
  /**
   * List of albums owned by the logged in user.
   */
  myAlbums: getMyAlbums_myAlbums[]
}
