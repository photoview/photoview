/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidbarGetPhotoShares
// ====================================================

export interface sidbarGetPhotoShares_media_shares {
  __typename: "ShareToken";
  id: string;
  token: string;
  /**
   * Whether or not a password is needed to access the share
   */
  hasPassword: boolean;
}

export interface sidbarGetPhotoShares_media {
  __typename: "Media";
  id: string;
  shares: sidbarGetPhotoShares_media_shares[];
}

export interface sidbarGetPhotoShares {
  /**
   * Get media by id, user must own the media or be admin.
   * If valid tokenCredentials are provided, the media may be retrived without further authentication
   */
  media: sidbarGetPhotoShares_media;
}

export interface sidbarGetPhotoSharesVariables {
  id: string;
}
