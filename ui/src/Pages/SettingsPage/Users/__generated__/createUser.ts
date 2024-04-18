/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createUser
// ====================================================

export interface createUser_createUser_role {
  __typename: 'Role'
  id: string
  name: string
}

export interface createUser_createUser {
  __typename: 'User'
  id: string
  username: string
  role: createUser_createUser_role
}

export interface createUser {
  /**
   * Create a new user
   */
  createUser: createUser_createUser
}

export interface createUserVariables {
  username: string
  roleId: string
}
