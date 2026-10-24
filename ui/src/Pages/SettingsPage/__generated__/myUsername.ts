/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: myUsername
// ====================================================

export interface myUsername_myUser {
  __typename: "User";
  id: string;
  username: string;
}

export interface myUsername {
  /**
   * Information about the currently logged in user
   */
  myUser: myUsername_myUser;
}
