/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidebarGetPhotoShares
// ====================================================

export interface sidebarGetPhotoShares_media_shares {
  __typename: 'ShareToken'
  id: string
  token: string
  /**
   * Whether or not a password is needed to access the share
   */
  hasPassword: boolean
}

export interface sidebarGetPhotoShares_media {
  __typename: 'Media'
  id: string
  /**
   * A list of share tokens pointing to this media, owned byt the logged in user
   */
  shares: sidebarGetPhotoShares_media_shares[]
}

export interface sidebarGetPhotoShares {
  /**
   * Get media by id, user must own the media or be admin.
   * If valid tokenCredentials are provided, the media may be retrived without further authentication
   */
  media: sidebarGetPhotoShares_media
}

export interface sidebarGetPhotoSharesVariables {
  id: string
}
