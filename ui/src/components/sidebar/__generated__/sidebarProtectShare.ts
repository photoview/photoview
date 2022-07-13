/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarProtectShare
// ====================================================

export interface sidebarProtectShare_protectShareToken {
  __typename: 'ShareToken'
  token: string
  /**
   * Whether or not a password is needed to access the share
   */
  hasPassword: boolean
}

export interface sidebarProtectShare {
  /**
   * Set a password for a token, if null is passed for the password argument, the password will be cleared
   */
  protectShareToken: sidebarProtectShare_protectShareToken
}

export interface sidebarProtectShareVariables {
  token: string
  password?: string | null
}
