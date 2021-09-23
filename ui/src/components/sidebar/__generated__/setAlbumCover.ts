/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: setAlbumCover
// ====================================================

export interface setAlbumCover_setAlbumCover_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface setAlbumCover_setAlbumCover_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: setAlbumCover_setAlbumCover_thumbnail_thumbnail | null
}

export interface setAlbumCover_setAlbumCover {
  __typename: 'Album'
  id: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: setAlbumCover_setAlbumCover_thumbnail | null
}

export interface setAlbumCover {
  /**
   * Assign a cover photo to an album
   */
  setAlbumCover: setAlbumCover_setAlbumCover
}

export interface setAlbumCoverVariables {
  coverID: string
}
