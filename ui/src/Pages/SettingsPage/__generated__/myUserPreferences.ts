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
   * The maximum number of albums/media to show per category in search results.
   * `null` uses the server default, `0` means no limit (show all results).
   */
  searchResultLimit: number | null;
  /**
   * Whether the album tree sidebar is shown in the UI.
   * `null` uses the default, which is to show it.
   */
  showAlbumTree: boolean | null;
  /**
   * Whether personally-hidden albums are shown (dimmed, with a click-to-unhide
   * affordance) instead of being excluded from navigation and search.
   * `null` uses the default, which is to exclude them.
   */
  showHiddenAlbums: boolean | null;
}

export interface myUserPreferences {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: myUserPreferences_myUserPreferences;
}
