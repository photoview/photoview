/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumTreeActivePathQuery
// ====================================================

export interface albumTreeActivePathQuery_album_path {
  __typename: "Album";
  id: string;
}

export interface albumTreeActivePathQuery_album {
  __typename: "Album";
  id: string;
  /**
   * A breadcrumb list of all parent albums down to this one
   */
  path: albumTreeActivePathQuery_album_path[];
}

export interface albumTreeActivePathQuery {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: albumTreeActivePathQuery_album;
}

export interface albumTreeActivePathQueryVariables {
  id: string;
}
