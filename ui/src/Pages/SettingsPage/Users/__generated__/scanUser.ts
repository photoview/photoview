/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: scanUser
// ====================================================

export interface scanUser_scanUser {
  __typename: 'ScannerResult'
  success: boolean
}

export interface scanUser {
  /**
   * Scan a single user for new media
   */
  scanUser: scanUser_scanUser
}

export interface scanUserVariables {
  userId: string
}
