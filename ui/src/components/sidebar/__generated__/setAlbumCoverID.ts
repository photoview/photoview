/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: setAlbumCoverID
// ====================================================

export interface setAlbumCoverID_setAlbumCoverID {
  __typename: 'Album'
  id: string
  coverID: string
}

export interface setAlbumCoverID {
  /**
   * Assign a cover image to an album, set coverID to -1 to remove the current one
   */
  setAlbumCoverID: setAlbumCoverID_setAlbumCoverID
}

export interface setAlbumCoverIDVariables {
  albumID: string
  coverID: string
}
