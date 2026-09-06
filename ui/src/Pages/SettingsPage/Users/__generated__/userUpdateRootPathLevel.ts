/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AlbumPermissionLevel } from "./../../../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: userUpdateRootPathLevel
// ====================================================

export interface userUpdateRootPathLevel_userUpdateRootPathLevel {
  __typename: "Album";
  id: string;
}

export interface userUpdateRootPathLevel {
  /**
   * Change the permission level a user holds on one of their existing root albums.
   */
  userUpdateRootPathLevel: userUpdateRootPathLevel_userUpdateRootPathLevel | null;
}

export interface userUpdateRootPathLevelVariables {
  id: string;
  albumId: string;
  level: AlbumPermissionLevel;
}
