/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { LanguageTranslation } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: myUserPreferences
// ====================================================

export interface myUserPreferences_myUserPreferences {
  __typename: "UserPreferences";
  id: string;
  language: LanguageTranslation | null;
  /**
   * How many results a search returns per category. 0 means as many as the server allows (currently 1000). Unset falls back to the server default.
   */
  searchResultLimit: number | null;
}

export interface myUserPreferences {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: myUserPreferences_myUserPreferences;
}
