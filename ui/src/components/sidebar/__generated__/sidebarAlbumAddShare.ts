/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: sidebarAlbumAddShare
// ====================================================

export interface sidebarAlbumAddShare_shareAlbum {
  __typename: 'ShareToken'
  token: string
}

export interface sidebarAlbumAddShare {
  /**
   * Generate share token for album
   */
  shareAlbum: sidebarAlbumAddShare_shareAlbum
}

export interface sidebarAlbumAddShareVariables {
  id: string
  password?: string | null
  expire?: Time | null
}
