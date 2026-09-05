/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: moveAlbum
// ====================================================

export interface moveAlbum_moveAlbum {
  __typename: 'Album'
  id: string
}

export interface moveAlbum {
  /**
   * Move a folder (and everything inside it) into a new parent album, both on
   * disk and in the database. The caller must be an admin, or must own both
   * `albumId` and `newParentAlbumId` and have the `canUpload` permission.
   * Root albums cannot be moved, and an album cannot be moved into itself or
   * one of its own sub-albums.
   */
  moveAlbum: moveAlbum_moveAlbum
}

export interface moveAlbumVariables {
  albumId: string
  newParentAlbumId: string
}
