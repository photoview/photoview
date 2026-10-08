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
   * How many results a search returns per category. 0 means as many as the server allows (currently 1000). Unset falls back to the server default.
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
