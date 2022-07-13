/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: scanAllMutation
// ====================================================

export interface scanAllMutation_scanAll {
  __typename: 'ScannerResult'
  success: boolean
  message: string | null
}

export interface scanAllMutation {
  /**
   * Scan all users for new media
   */
  scanAll: scanAllMutation_scanAll
}
