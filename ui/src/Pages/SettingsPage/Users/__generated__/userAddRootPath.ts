/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: userAddRootPath
// ====================================================

export interface userAddRootPath_userAddRootPath {
  __typename: 'Album'
  id: string
}

export interface userAddRootPath {
  /**
   * Add a root path from where to look for media for the given user, specified by their user id.
   */
  userAddRootPath: userAddRootPath_userAddRootPath | null
}

export interface userAddRootPathVariables {
  id: string
  rootPath: string
}
