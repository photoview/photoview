/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: CheckInitialSetup
// ====================================================

export interface CheckInitialSetup_siteInfo {
  __typename: 'SiteInfo'
  /**
   * Whether or not the initial setup wizard should be shown
   */
  initialSetup: boolean
}

export interface CheckInitialSetup {
  siteInfo: CheckInitialSetup_siteInfo
}
