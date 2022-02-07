/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: settingsUsersQuery
// ====================================================

export interface settingsUsersQuery_user_rootAlbums {
  __typename: 'Album'
  id: string
  /**
   * The path on the filesystem of the server, where this album is located
   */
  filePath: string
}

export interface settingsUsersQuery_user {
  __typename: 'User'
  id: string
  username: string
  /**
   * Whether or not the user has admin privileges
   */
  admin: boolean
  /**
   * Top level albums owned by this user
   */
  rootAlbums: settingsUsersQuery_user_rootAlbums[]
}

export interface settingsUsersQuery {
  /**
   * List of registered users, must be admin to call
   */
  user: settingsUsersQuery_user[]
}
