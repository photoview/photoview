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
  admin: boolean;
}

export interface createUser {
  createUser: createUser_createUser | null;
}

export interface createUserVariables {
  username: string;
  admin: boolean;
}
