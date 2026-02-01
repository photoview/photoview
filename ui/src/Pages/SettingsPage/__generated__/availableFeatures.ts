/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: availableFeatures
// ====================================================

export interface availableFeatures_siteInfo {
  __typename: "SiteInfo";
  /**
   * Whether or not face detection is enabled and working
   */
  faceDetectionEnabled: boolean;
}

export interface availableFeatures {
  /**
   * Get the mapbox api token, returns null if mapbox is not enabled
   */
  mapboxToken: string | null;
  siteInfo: availableFeatures_siteInfo;
}
