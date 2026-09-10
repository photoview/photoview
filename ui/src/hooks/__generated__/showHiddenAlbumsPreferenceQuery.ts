/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: showHiddenAlbumsPreferenceQuery
// ====================================================

export interface showHiddenAlbumsPreferenceQuery_myUserPreferences {
  __typename: 'UserPreferences'
  id: string
  /**
   * Whether personally-hidden albums are shown (dimmed, with a click-to-unhide
   * affordance) instead of being excluded from navigation and search.
   * `null` uses the default, which is to exclude them.
   */
  showHiddenAlbums: boolean | null
}

export interface showHiddenAlbumsPreferenceQuery {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: showHiddenAlbumsPreferenceQuery_myUserPreferences
}
