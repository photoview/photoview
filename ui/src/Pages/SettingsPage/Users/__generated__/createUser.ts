/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createUser
// ====================================================

export interface createUser_createUser {
  __typename: "User";
  id: string;
  username: string;
  /**
   * Whether or not the user has admin privileges
   */
  admin: boolean;
  /**
   * Whether the user may share albums they own with other users. Admins can always share regardless of this flag.
   */
  canShare: boolean;
}

export interface createUser {
  /**
   * Create a new user
   */
  createUser: createUser_createUser;
}

export interface createUserVariables {
  username: string;
  admin: boolean;
  canShare?: boolean | null;
}
