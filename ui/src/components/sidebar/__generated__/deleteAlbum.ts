/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: deleteAlbum
// ====================================================

export interface deleteAlbum {
  /**
   * Move a folder (and everything inside it) to a hidden trash folder next to
   * its parent, and remove it from the library. This does not permanently
   * delete the files - an administrator can still recover them directly from
   * the filesystem. The caller must be an admin, or must own `albumId` and
   * have the `canUpload` permission. Root albums cannot be deleted this way.
   */
  deleteAlbum: boolean
}

export interface deleteAlbumVariables {
  albumId: string
}
