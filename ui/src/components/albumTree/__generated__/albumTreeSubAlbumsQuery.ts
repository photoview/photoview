/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumTreeSubAlbumsQuery
// ====================================================

export interface albumTreeSubAlbumsQuery_album_subAlbums {
  __typename: "Album";
  id: string;
  title: string;
}

export interface albumTreeSubAlbumsQuery_album {
  __typename: "Album";
  id: string;
  /**
   * The albums contained in this album
   */
  subAlbums: albumTreeSubAlbumsQuery_album_subAlbums[];
}

export interface albumTreeSubAlbumsQuery {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: albumTreeSubAlbumsQuery_album;
}

export interface albumTreeSubAlbumsQueryVariables {
  id: string;
}
