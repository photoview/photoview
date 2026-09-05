/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarSetExpireShare
// ====================================================

export interface sidebarSetExpireShare_setExpireShareToken {
  __typename: 'ShareToken'
  token: string
}

export interface sidebarSetExpireShare {
  /**
   * Set a Expiration Time for a token
   */
  setExpireShareToken: sidebarSetExpireShare_setExpireShareToken
}

export interface sidebarSetExpireShareVariables {
  token: string
  expire?: Time | null
}
