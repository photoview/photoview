/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarSetShareTokenLabel
// ====================================================

export interface sidebarSetShareTokenLabel_setShareTokenLabel {
  __typename: 'ShareToken'
  token: string
  /**
   * Optional label visible to the owner or an administrator
   */
  label: string | null
}

export interface sidebarSetShareTokenLabel {
  /**
   * Set an optional label for a share token
   */
  setShareTokenLabel: sidebarSetShareTokenLabel_setShareTokenLabel
}

export interface sidebarSetShareTokenLabelVariables {
  token: string
  label?: string | null
}
