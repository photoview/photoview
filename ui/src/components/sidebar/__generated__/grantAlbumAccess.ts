/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AlbumPermissionLevel } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: grantAlbumAccess
// ====================================================

export interface grantAlbumAccess_grantAlbumAccess_user {
  __typename: "User";
  id: string;
  username: string;
}

export interface grantAlbumAccess_grantAlbumAccess {
  __typename: "AlbumPermission";
  user: grantAlbumAccess_grantAlbumAccess_user;
  level: AlbumPermissionLevel;
}

export interface grantAlbumAccess {
  /**
   * Grant another user access to one of your own folders at the given level.
   * Only the folder's owner (the admin-granted holder, not someone who
   * received it via another share) may do this, and only up to their own
   * level.
   */
  grantAlbumAccess: grantAlbumAccess_grantAlbumAccess;
}

export interface grantAlbumAccessVariables {
  albumId: string;
  userId: string;
  level: AlbumPermissionLevel;
}
