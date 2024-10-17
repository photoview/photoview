/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarToggleShareDownload
// ====================================================

export interface sidebarToggleShareDownload_toggleShareDownload {
  __typename: 'ShareToken'
  token: string
  /**
   * Whether to allow downloading this share
   */
  allowDownload: boolean
}

export interface sidebarToggleShareDownload {
  /**
   * Mark a share as downloadable by non-authenticated users
   */
  toggleShareDownload: sidebarToggleShareDownload_toggleShareDownload
}

export interface sidebarToggleShareDownloadVariables {
  token: string
  allowDownload: boolean
}
