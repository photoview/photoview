/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: deleteUser
// ====================================================

export interface deleteUser_deleteUser {
  __typename: 'User'
  id: string
  username: string
}

export interface deleteUser {
  /**
   * Delete an existing user
   */
  deleteUser: deleteUser_deleteUser
}

export interface deleteUserVariables {
  id: string
}
