/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: setAlbumCover
// ====================================================

export interface setAlbumCover_setAlbumCover {
  __typename: 'Album'
  coverID: number
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
