/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: changeUserPassword
// ====================================================

export interface changeUserPassword_updateUser {
  __typename: 'User'
  id: string
}

export interface changeUserPassword {
  /**
   * Update a user, fields left as `null` will not be changed
   */
  updateUser: changeUserPassword_updateUser
}

export interface changeUserPasswordVariables {
  userId: string
  password: string
}
