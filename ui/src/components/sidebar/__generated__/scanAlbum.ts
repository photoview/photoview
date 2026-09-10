/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: scanAlbum
// ====================================================

export interface scanAlbum_scanAlbum {
  __typename: "ScannerResult";
  success: boolean;
}

export interface scanAlbum {
  /**
   * Recursively scan a single album and its sub-albums for new media, without
   * scanning the rest of the library. Caller must be an admin, or hold at
   * least UPLOAD-level access on albumId.
   */
  scanAlbum: scanAlbum_scanAlbum;
}

export interface scanAlbumVariables {
  albumId: string;
}
