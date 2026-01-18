/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarSetExpierShare
// ====================================================

export interface sidebarSetExpierShare_setExpireShareToken {
  __typename: "ShareToken";
  token: string;
}

export interface sidebarSetExpierShare {
  /**
   * Set a Expiration Time for a token
   */
  setExpireShareToken: sidebarSetExpierShare_setExpireShareToken;
}

export interface sidebarSetExpierShareVariables {
  token: string;
  expire?: Time | null;
}
