/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: resetAlbumCover
// ====================================================

export interface resetAlbumCover_resetAlbumCover_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface resetAlbumCover_resetAlbumCover_thumbnail {
  __typename: 'Media'
  id: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: resetAlbumCover_resetAlbumCover_thumbnail_thumbnail | null
}

export interface resetAlbumCover_resetAlbumCover {
  __typename: 'Album'
  id: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: resetAlbumCover_resetAlbumCover_thumbnail | null
}

export interface resetAlbumCover {
  /**
   * Reset the assigned cover photo for an album
   */
  resetAlbumCover: resetAlbumCover_resetAlbumCover
}

export interface resetAlbumCoverVariables {
  albumID: string
}
