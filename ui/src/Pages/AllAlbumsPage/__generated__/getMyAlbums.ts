/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { OrderDirection } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: getMyAlbums
// ====================================================

export interface getMyAlbums_myAlbums_thumbnail_thumbnail {
  __typename: "MediaURL";
  /**
   * URL for previewing the image
   */
  url: string;
}

export interface getMyAlbums_myAlbums_thumbnail {
  __typename: "Media";
  id: string;
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: getMyAlbums_myAlbums_thumbnail_thumbnail | null;
}

export interface getMyAlbums_myAlbums {
  __typename: "Album";
  id: string;
  title: string;
  /**
   * Whether the currently logged in user has personally hidden this album from their own navigation. Never affects other users.
   */
  viewerHidden: boolean;
  /**
   * ID of the album which contains this album, or null if this is a root album.
   * Unlike parentAlbum, this doesn't require the parent association to be
   * preloaded, so it's always accurate.
   */
  parentAlbumId: string | null;
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: getMyAlbums_myAlbums_thumbnail | null;
}

export interface getMyAlbums {
  /**
   * List of albums owned by the logged in user.
   */
  myAlbums: getMyAlbums_myAlbums[];
}

export interface getMyAlbumsVariables {
  orderBy?: string | null;
  orderDirection?: OrderDirection | null;
  showHidden?: boolean | null;
}
