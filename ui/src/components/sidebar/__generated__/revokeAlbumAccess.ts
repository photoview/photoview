/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: revokeAlbumAccess
// ====================================================

export interface revokeAlbumAccess {
  /**
   * Revoke a share you previously created.
   */
  revokeAlbumAccess: boolean;
}

export interface revokeAlbumAccessVariables {
  albumId: string;
  userId: string;
}
