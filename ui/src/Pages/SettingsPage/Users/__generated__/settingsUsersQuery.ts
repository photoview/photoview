/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AlbumPermissionLevel } from "./../../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: settingsUsersQuery
// ====================================================

export interface settingsUsersQuery_user_rootAlbums_permissions_user {
  __typename: "User";
  id: string;
}

export interface settingsUsersQuery_user_rootAlbums_permissions {
  __typename: "AlbumPermission";
  user: settingsUsersQuery_user_rootAlbums_permissions_user;
  level: AlbumPermissionLevel;
}

export interface settingsUsersQuery_user_rootAlbums {
  __typename: "Album";
  id: string;
  /**
   * The path on the filesystem of the server, where this album is located
   */
  filePath: string;
  /**
   * Users with an explicit access grant directly on this album. Only populated for the album's owner (or an admin) - null otherwise.
   */
  permissions: settingsUsersQuery_user_rootAlbums_permissions[] | null;
}

export interface settingsUsersQuery_user {
  __typename: "User";
  id: string;
  username: string;
  /**
   * Whether or not the user has admin privileges
   */
  admin: boolean;
  /**
   * Whether the user may share albums they own with other users. Admins can always share regardless of this flag.
   */
  canShare: boolean;
  /**
   * Top level albums owned by this user
   */
  rootAlbums: settingsUsersQuery_user_rootAlbums[];
}

export interface settingsUsersQuery {
  /**
   * List of registered users, must be admin to call
   */
  user: settingsUsersQuery_user[];
}
