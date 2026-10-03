/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: layoutAlbumTreePreferenceQuery
// ====================================================

export interface layoutAlbumTreePreferenceQuery_myUserPreferences {
  __typename: "UserPreferences";
  id: string;
  /**
   * Whether to show the album tree next to the gallery. Off unless the user turns it on.
   */
  showAlbumTree: boolean | null;
}

export interface layoutAlbumTreePreferenceQuery {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: layoutAlbumTreePreferenceQuery_myUserPreferences;
}
