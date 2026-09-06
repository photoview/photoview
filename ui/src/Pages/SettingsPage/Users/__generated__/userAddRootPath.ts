/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AlbumPermissionLevel } from "./../../../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: userAddRootPath
// ====================================================

export interface userAddRootPath_userAddRootPath {
  __typename: "Album";
  id: string;
}

export interface userAddRootPath {
  /**
   * Add a root path from where to look for media for the given user, specified
   * by their user id, at the given permission level (defaults to READ if
   * omitted).
   */
  userAddRootPath: userAddRootPath_userAddRootPath | null;
}

export interface userAddRootPathVariables {
  id: string;
  rootPath: string;
  level?: AlbumPermissionLevel | null;
}
