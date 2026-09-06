/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumTreeRootQuery
// ====================================================

export interface albumTreeRootQuery_myAlbums {
  __typename: "Album";
  id: string;
  title: string;
  /**
   * Whether the currently logged in user has personally hidden this album from their own navigation. Never affects other users.
   */
  viewerHidden: boolean;
  /**
   * ID of the album which contains this album, or null if this is a root album.
   * Unlike parentAlbum, this doesn't require the parent association to be
   * preloaded, so it's always accurate.
   */
  parentAlbumId: string | null;
}

export interface albumTreeRootQuery {
  /**
   * List of albums owned by the logged in user.
   */
  myAlbums: albumTreeRootQuery_myAlbums[];
}

export interface albumTreeRootQueryVariables {
  showHidden?: boolean | null;
}
