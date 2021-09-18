/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: resetAlbumCover
// ====================================================

export interface resetAlbumCover_resetAlbumCover {
  __typename: 'Album'
  id: string
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
