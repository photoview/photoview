/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: getAlbumSidebar
// ====================================================

export interface getAlbumSidebar_album {
  __typename: 'Album'
  id: string
  title: string
}

export interface getAlbumSidebar {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: getAlbumSidebar_album
}

export interface getAlbumSidebarVariables {
  id: string
}
