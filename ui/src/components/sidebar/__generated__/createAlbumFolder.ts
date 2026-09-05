/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createAlbumFolder
// ====================================================

export interface createAlbumFolder_createAlbumFolder {
  __typename: 'Album'
  id: string
  title: string
}

export interface createAlbumFolder {
  /**
   * Create a new, empty sub-folder on disk inside the given album.
   * The caller must be an admin, or must own `parentAlbumId` and have the
   * `canUpload` permission.
   */
  createAlbumFolder: createAlbumFolder_createAlbumFolder
}

export interface createAlbumFolderVariables {
  parentAlbumId: string
  name: string
}
