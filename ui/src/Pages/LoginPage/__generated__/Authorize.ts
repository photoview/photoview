/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: Authorize
// ====================================================

export interface Authorize_authorizeUser {
  __typename: "AuthorizeResult";
  success: boolean;
  status: string;
  token: string | null;
}

export interface Authorize {
  authorizeUser: Authorize_authorizeUser;
}

export interface AuthorizeVariables {
  username: string;
  password: string;
}
