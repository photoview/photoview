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
}

export interface albumTreeRootQuery {
  /**
   * List of albums owned by the logged in user.
   */
  myAlbums: albumTreeRootQuery_myAlbums[];
}
