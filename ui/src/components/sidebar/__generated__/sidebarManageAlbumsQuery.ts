/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: sidebarManageAlbumsQuery
// ====================================================

export interface sidebarManageAlbumsQuery_myAlbums {
  __typename: 'Album'
  id: string
  title: string
}

export interface sidebarManageAlbumsQuery {
  /**
   * List of albums owned by the logged in user.
   */
  myAlbums: sidebarManageAlbumsQuery_myAlbums[]
}
