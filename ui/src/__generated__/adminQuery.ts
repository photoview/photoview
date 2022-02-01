/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: adminQuery
// ====================================================

export interface adminQuery_myUser {
  __typename: 'User'
  admin: boolean
}

export interface adminQuery {
  /**
   * Information about the currently logged in user
   */
  myUser: adminQuery_myUser
}
