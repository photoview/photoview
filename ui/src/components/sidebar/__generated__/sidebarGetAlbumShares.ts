/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidebarGetAlbumShares
// ====================================================

export interface sidebarGetAlbumShares_album_shares {
  __typename: 'ShareToken'
  id: string
  token: string
  /**
   * Whether or not a password is needed to access the share
   */
  hasPassword: boolean
}

export interface sidebarGetAlbumShares_album {
  __typename: 'Album'
  id: string
  /**
   * A list of share tokens pointing to this album, owned by the logged in user
   */
  shares: sidebarGetAlbumShares_album_shares[]
}

export interface sidebarGetAlbumShares {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: sidebarGetAlbumShares_album
}

export interface sidebarGetAlbumSharesVariables {
  id: string
}
