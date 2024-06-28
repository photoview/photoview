/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateUser
// ====================================================

export interface updateUser_updateUser_role {
  __typename: 'Role'
  id: string
  name: string
}

export interface updateUser_updateUser {
  __typename: 'User'
  id: string
  username: string
  admin: boolean
  role: updateUser_updateUser_role
}

export interface updateUser {
  /**
   * Update a user, fields left as `null` will not be changed
   */
  updateUser: updateUser_updateUser
}

export interface updateUserVariables {
  id: string
  username?: string | null
  roleId: string
}
