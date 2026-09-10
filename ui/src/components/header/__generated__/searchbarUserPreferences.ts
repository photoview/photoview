/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: searchbarUserPreferences
// ====================================================

export interface searchbarUserPreferences_myUserPreferences {
  __typename: "UserPreferences";
  id: string;
  /**
   * The maximum number of albums/media to show per category in search results.
   * `null` uses the server default, `0` means no limit (show all results).
   */
  searchResultLimit: number | null;
}

export interface searchbarUserPreferences {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: searchbarUserPreferences_myUserPreferences;
}
