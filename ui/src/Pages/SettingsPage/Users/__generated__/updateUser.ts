/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateUser
// ====================================================

export interface updateUser_updateUser {
  __typename: 'User'
  id: string
  username: string
  /**
   * Whether or not the user has admin privileges
   */
  admin: boolean
  /**
   * Whether or not the user may create folders and upload media into albums they own
   */
  canUpload: boolean
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
  admin?: boolean | null
  canUpload?: boolean | null
}
