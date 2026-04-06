/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: mapStylesQuery
// ====================================================

export interface mapStylesQuery_siteInfo {
  __typename: "SiteInfo";
  /**
   * Custom map tile style URL for light mode, null means use the built-in style
   */
  mapStyleLight: string | null;
  /**
   * Custom map tile style URL for dark mode, null means use the built-in style
   */
  mapStyleDark: string | null;
}

export interface mapStylesQuery {
  siteInfo: mapStylesQuery_siteInfo;
}
