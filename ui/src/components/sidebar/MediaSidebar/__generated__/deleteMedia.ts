/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: deleteMedia
// ====================================================

export interface deleteMedia {
  /**
   * Move a media file (and its sidecar file, if any) to a hidden trash
   * folder next to its album, and remove it from the library. Does not
   * permanently delete the files - an administrator can still recover them
   * directly from the filesystem. The caller must be an admin, or hold at
   * least UPLOAD-level access on the media's album.
   */
  deleteMedia: boolean;
}

export interface deleteMediaVariables {
  mediaId: string;
}
