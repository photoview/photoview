/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { MediaType } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: searchPageQuery
// ====================================================

export interface searchPageQuery_search_albums_thumbnail_thumbnail {
  __typename: "MediaURL";
  /**
   * URL for previewing the image
   */
  url: string;
}

export interface searchPageQuery_search_albums_thumbnail {
  __typename: "Media";
  id: string;
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: searchPageQuery_search_albums_thumbnail_thumbnail | null;
}

export interface searchPageQuery_search_albums {
  __typename: "Album";
  id: string;
  title: string;
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: searchPageQuery_search_albums_thumbnail | null;
}

export interface searchPageQuery_search_media_thumbnail {
  __typename: "MediaURL";
  /**
   * URL for previewing the image
   */
  url: string;
  /**
   * Width of the image in pixels
   */
  width: number;
  /**
   * Height of the image in pixels
   */
  height: number;
}

export interface searchPageQuery_search_media_highRes {
  __typename: "MediaURL";
  /**
   * URL for previewing the image
   */
  url: string;
}

export interface searchPageQuery_search_media_videoWeb {
  __typename: "MediaURL";
  /**
   * URL for previewing the image
   */
  url: string;
}

export interface searchPageQuery_search_media_album {
  __typename: "Album";
  id: string;
  title: string;
}

export interface searchPageQuery_search_media {
  __typename: "Media";
  id: string;
  type: MediaType;
  /**
   * A short string that can be used to generate a blured version of the media, to show while the original is loading
   */
  blurhash: string | null;
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: searchPageQuery_search_media_thumbnail | null;
  /**
   * URL to display the photo in full resolution, will be null for videos
   */
  highRes: searchPageQuery_search_media_highRes | null;
  /**
   * URL to get the video in a web format that can be played in the browser, will be null for photos
   */
  videoWeb: searchPageQuery_search_media_videoWeb | null;
  favorite: boolean;
  /**
   * The album that holds the media
   */
  album: searchPageQuery_search_media_album;
}

export interface searchPageQuery_search {
  __typename: "SearchResult";
  /**
   * A list of albums that matched the query
   */
  albums: searchPageQuery_search_albums[];
  /**
   * A list of media that matched the query
   */
  media: searchPageQuery_search_media[];
}

export interface searchPageQuery {
  /**
   * Perform a search query on the contents of the media library
   */
  search: searchPageQuery_search;
}

export interface searchPageQueryVariables {
  query: string;
  limit?: number | null;
}
