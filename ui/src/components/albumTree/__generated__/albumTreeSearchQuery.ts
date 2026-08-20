/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumTreeSearchQuery
// ====================================================

export interface albumTreeSearchQuery_search_albums_path {
  __typename: "Album";
  id: string;
}

export interface albumTreeSearchQuery_search_albums {
  __typename: "Album";
  id: string;
  /**
   * A breadcrumb list of all parent albums down to this one
   */
  path: albumTreeSearchQuery_search_albums_path[];
}

export interface albumTreeSearchQuery_search {
  __typename: "SearchResult";
  /**
   * A list of albums that matched the query
   */
  albums: albumTreeSearchQuery_search_albums[];
}

export interface albumTreeSearchQuery {
  /**
   * Perform a search query on the contents of the media library
   */
  search: albumTreeSearchQuery_search;
}

export interface albumTreeSearchQueryVariables {
  query: string;
}
