/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumPathQuery
// ====================================================

export interface albumPathQuery_album_path {
  __typename: 'Album'
  id: string
  title: string
}

export interface albumPathQuery_album {
  __typename: 'Album'
  id: string
  path: albumPathQuery_album_path[]
}

export interface albumPathQuery {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: albumPathQuery_album
}

export interface albumPathQueryVariables {
  id: string
}
