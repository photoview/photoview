/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: AuthorizeGoogleOAuth
// ====================================================

export interface AuthorizeGoogleOAuth_authorizeGoogleOAuth {
  __typename: "AuthorizeResult";
  success: boolean;
  /**
   * A textual status message describing the result, can be used to show an error message when `success` is false
   */
  status: string;
  /**
   * An access token used to authenticate new API requests as the newly authorized user. Is present when success is true
   */
  token: string | null;
}

export interface AuthorizeGoogleOAuth {
  /**
   * Authorize a user using a Google OAuth JWT, returns an access token on success
   */
  authorizeGoogleOAuth: AuthorizeGoogleOAuth_authorizeGoogleOAuth;
}

export interface AuthorizeGoogleOAuthVariables {
  jwt: string;
}
