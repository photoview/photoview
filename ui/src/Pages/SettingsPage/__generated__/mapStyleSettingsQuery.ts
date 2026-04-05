/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: mapStyleSettingsQuery
// ====================================================

export interface mapStyleSettingsQuery_siteInfo {
  __typename: "SiteInfo";
  /**
   * Map tile style URL for light mode
   */
  mapStyleLight: string;
  /**
   * Map tile style URL for dark mode
   */
  mapStyleDark: string;
}

export interface mapStyleSettingsQuery {
  siteInfo: mapStyleSettingsQuery_siteInfo;
}
