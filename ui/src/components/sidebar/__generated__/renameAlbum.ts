/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: renameAlbum
// ====================================================

export interface renameAlbum_renameAlbum {
  __typename: 'Album'
  id: string
  title: string
}

export interface renameAlbum {
  /**
   * Rename a folder on disk and in the database, keeping it in the same
   * parent album. The caller must be an admin, or hold at least
   * DELETE-level access on albumId. Root albums cannot be renamed.
   */
  renameAlbum: renameAlbum_renameAlbum
}

export interface renameAlbumVariables {
  albumId: string
  newName: string
}
