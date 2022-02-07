/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { LanguageTranslation } from './globalTypes'

// ====================================================
// GraphQL query operation: siteTranslation
// ====================================================

export interface siteTranslation_myUserPreferences {
  __typename: 'UserPreferences'
  id: string
  language: LanguageTranslation | null
}

export interface siteTranslation {
  /**
   * User preferences for the logged in user
   */
  myUserPreferences: siteTranslation_myUserPreferences
}
