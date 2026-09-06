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
   * Whether the currently logged in user is the owner of this album (an admin-configured grant, not received via another user's share) and may share it with other users
   */
  viewerIsOwner: boolean;
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
