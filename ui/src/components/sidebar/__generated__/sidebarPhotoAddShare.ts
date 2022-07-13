/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarPhotoAddShare
// ====================================================

export interface sidebarPhotoAddShare_shareMedia {
  __typename: 'ShareToken'
  token: string
}

export interface sidebarPhotoAddShare {
  /**
   * Generate share token for media
   */
  shareMedia: sidebarPhotoAddShare_shareMedia
}

export interface sidebarPhotoAddShareVariables {
  id: string
  password?: string | null
  expire?: Time | null
}
