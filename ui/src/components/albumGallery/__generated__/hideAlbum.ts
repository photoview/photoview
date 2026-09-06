/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: hideAlbum
// ====================================================

export interface hideAlbum_hideAlbum {
  __typename: "Album";
  id: string;
  /**
   * Whether the currently logged in user has personally hidden this album from their own navigation. Never affects other users.
   */
  viewerHidden: boolean;
}

export interface hideAlbum {
  /**
   * Hide or unhide an album from the logged-in user's own navigation. Purely personal - does not change any permission or other user's visibility.
   */
  hideAlbum: hideAlbum_hideAlbum;
}

export interface hideAlbumVariables {
  albumId: string;
  hidden: boolean;
}
