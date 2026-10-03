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
   * scanning the rest of the library. Sub-albums whose directories no longer
   * exist are removed. Caller must be an admin or own the album.
   */
  scanAlbum: scanAlbum_scanAlbum;
}

export interface scanAlbumVariables {
  albumId: string;
}
