/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarSetShareTokenName
// ====================================================

export interface sidebarSetShareTokenName_setShareTokenName {
  __typename: "ShareToken";
  token: string;
  /**
   * Optional name visible to the owner or an administrator
   */
  name: string | null;
}

export interface sidebarSetShareTokenName {
  /**
   * Set an optional name for a share token
   */
  setShareTokenName: sidebarSetShareTokenName_setShareTokenName;
}

export interface sidebarSetShareTokenNameVariables {
  token: string;
  name?: string | null;
}
