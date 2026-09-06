/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: renameMedia
// ====================================================

export interface renameMedia_renameMedia {
  __typename: "Media";
  id: string;
  title: string;
  /**
   * Local filepath for the media
   */
  path: string;
}

export interface renameMedia {
  /**
   * Rename a media file, on disk and in the library (and its sidecar file, if any). The caller must be an admin, or hold at least UPLOAD-level access on the media's album.
   */
  renameMedia: renameMedia_renameMedia;
}

export interface renameMediaVariables {
  mediaId: string;
  newName: string;
}
