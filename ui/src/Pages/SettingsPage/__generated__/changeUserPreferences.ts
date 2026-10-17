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
   * Whether to show the album tree next to the gallery. Off unless the user turns it on.
   */
  showAlbumTree: boolean | null;
}

export interface changeUserPreferences {
  /**
   * Change user preferences for the logged in user
   */
  changeUserPreferences: changeUserPreferences_changeUserPreferences;
}

export interface changeUserPreferencesVariables {
  language?: string | null;
  showAlbumTree?: boolean | null;
}
