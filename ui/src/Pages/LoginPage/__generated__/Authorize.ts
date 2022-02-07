/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: Authorize
// ====================================================

export interface Authorize_authorizeUser {
  __typename: 'AuthorizeResult'
  success: boolean
  /**
   * A textual status message describing the result, can be used to show an error message when `success` is false
   */
  status: string
  /**
   * An access token used to authenticate new API requests as the newly authorized user. Is present when success is true
   */
  token: string | null
}

export interface Authorize {
  /**
   * Authorizes a user and returns a token used to identify the new session
   */
  authorizeUser: Authorize_authorizeUser
}

export interface AuthorizeVariables {
  username: string
  password: string
}
