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
   * Whether the album tree sidebar is shown in the UI.
   * `null` uses the default, which is to show it.
   */
  showAlbumTree: boolean | null;
}

export interface layoutAlbumTreePreferenceQuery {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: layoutAlbumTreePreferenceQuery_myUserPreferences;
}
