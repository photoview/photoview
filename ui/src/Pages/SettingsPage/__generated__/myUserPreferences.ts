/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { LanguageTranslation } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL query operation: myUserPreferences
// ====================================================

export interface myUserPreferences_myUserPreferences {
  __typename: 'UserPreferences'
  id: string
  language: LanguageTranslation | null
}

export interface myUserPreferences {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: myUserPreferences_myUserPreferences
}
