/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: getAlbumSidebar
// ====================================================

export interface getAlbumSidebar_album {
  __typename: "Album";
  id: string;
  title: string;
  /**
   * Whether the currently logged in user may create folders in, or upload media into, this album
   */
  viewerCanUpload: boolean;
  /**
   * Whether the currently logged in user may delete or move this album
   */
  viewerCanDelete: boolean;
  /**
   * Whether the currently logged in user is the owner of this album (an admin-configured grant, not received via another user's share) and may share it with other users
   */
  viewerIsOwner: boolean;
  /**
   * ID of the album which contains this album, or null if this is a root album.
   * Unlike parentAlbum, this doesn't require the parent association to be
   * preloaded, so it's always accurate.
   */
  parentAlbumId: string | null;
}

export interface getAlbumSidebar {
  /**
   * Get album by id, user must own the album or be admin
   * If valid tokenCredentials are provided, the album may be retrived without further authentication
   */
  album: getAlbumSidebar_album;
}

export interface getAlbumSidebarVariables {
  id: string;
}
