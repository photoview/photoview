/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebareDeleteShare
// ====================================================

export interface sidebareDeleteShare_deleteShareToken {
  __typename: 'ShareToken'
  token: string
}

export interface sidebareDeleteShare {
  /**
   * Delete a share token by it's token value
   */
  deleteShareToken: sidebareDeleteShare_deleteShareToken
}

export interface sidebareDeleteShareVariables {
  token: string
}
