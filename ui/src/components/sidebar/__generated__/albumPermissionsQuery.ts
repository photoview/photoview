/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { AlbumPermissionLevel } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: albumPermissionsQuery
// ====================================================

export interface albumPermissionsQuery_shareableUsers {
  __typename: "User";
  id: string;
  username: string;
}

export interface albumPermissionsQuery_album_permissions_user {
  __typename: "User";
  id: string;
  username: string;
}

export interface albumPermissionsQuery_album_permissions {
  __typename: "AlbumPermission";
  user: albumPermissionsQuery_album_permissions_user;
  level: AlbumPermissionLevel;
}

export interface albumPermissionsQuery_album {
  __typename: "Album";
  id: string;
  /**
   * Users with an explicit access grant directly on this album. Only populated for the album's owner (or an admin) - null otherwise.
   */
  permissions: albumPermissionsQuery_album_permissions[] | null;
}

export interface albumPermissionsQuery {
  /**
   * Other users this account may pick as a share target. Username/id only.
   */
  shareableUsers: albumPermissionsQuery_shareableUsers[];
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: albumPermissionsQuery_album;
}

export interface albumPermissionsQueryVariables {
  albumId: string;
}
