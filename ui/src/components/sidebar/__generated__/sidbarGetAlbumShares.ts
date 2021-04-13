/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidbarGetAlbumShares
// ====================================================

export interface sidbarGetAlbumShares_album_shares {
  __typename: "ShareToken";
  id: string;
  token: string;
  /**
   * Whether or not a password is needed to access the share
   */
  hasPassword: boolean;
}

export interface sidbarGetAlbumShares_album {
  __typename: "Album";
  id: string;
  shares: (sidbarGetAlbumShares_album_shares | null)[] | null;
}

export interface sidbarGetAlbumShares {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: sidbarGetAlbumShares_album;
}

export interface sidbarGetAlbumSharesVariables {
  id: string;
}
