/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: changeUserPassword
// ====================================================

export interface changeUserPassword_updateUser {
  __typename: "User";
  id: string;
}

export interface changeUserPassword {
  updateUser: changeUserPassword_updateUser | null;
}

export interface changeUserPasswordVariables {
  userId: string;
  password: string;
}
