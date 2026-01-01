/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: indexPagePreferences
// ====================================================

export interface indexPagePreferences_myUserPreferences {
  __typename: "UserPreferences";
  id: string;
  defaultLandingPage: string | null;
}

export interface indexPagePreferences {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: indexPagePreferences_myUserPreferences;
}
