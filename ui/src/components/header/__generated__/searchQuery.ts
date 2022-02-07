/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: searchQuery
// ====================================================

export interface searchQuery_search_albums_thumbnail_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface searchQuery_search_albums_thumbnail {
  __typename: 'Media'
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: searchQuery_search_albums_thumbnail_thumbnail | null
}

export interface searchQuery_search_albums {
  __typename: 'Album'
  id: string
  title: string
  /**
   * An image in this album used for previewing this album
   */
  thumbnail: searchQuery_search_albums_thumbnail | null
}

export interface searchQuery_search_media_thumbnail {
  __typename: 'MediaURL'
  /**
   * URL for previewing the image
   */
  url: string
}

export interface searchQuery_search_media_album {
  __typename: 'Album'
  id: string
}

export interface searchQuery_search_media {
  __typename: 'Media'
  id: string
  title: string
  /**
   * URL to display the media in a smaller resolution
   */
  thumbnail: searchQuery_search_media_thumbnail | null
  /**
   * The album that holds the media
   */
  album: searchQuery_search_media_album
}

export interface searchQuery_search {
  __typename: 'SearchResult'
  /**
   * The string that was searched for
   */
  query: string
  /**
   * A list of albums that matched the query
   */
  albums: searchQuery_search_albums[]
  /**
   * A list of media that matched the query
   */
  media: searchQuery_search_media[]
}

export interface searchQuery {
  /**
   * Perform a search query on the contents of the media library
   */
  search: searchQuery_search
}

export interface searchQueryVariables {
  query: string
}
