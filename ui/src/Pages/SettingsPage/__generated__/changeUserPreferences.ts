/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { LanguageTranslation } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: changeUserPreferences
// ====================================================

export interface changeUserPreferences_changeUserPreferences {
  __typename: "UserPreferences";
  id: string;
  language: LanguageTranslation | null;
  /**
   * The maximum number of albums/media to show per category in search results.
   * `null` uses the server default, `0` means no limit (show all results).
   */
  searchResultLimit: number | null;
}

export interface changeUserPreferences {
  /**
   * Change user preferences for the logged in user
   */
  changeUserPreferences: changeUserPreferences_changeUserPreferences;
}

export interface changeUserPreferencesVariables {
  language?: string | null;
  searchResultLimit?: number | null;
}
