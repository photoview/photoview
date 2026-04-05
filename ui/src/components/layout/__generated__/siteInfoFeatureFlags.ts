/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: siteInfoFeatureFlags
// ====================================================

export interface siteInfoFeatureFlags_siteInfo {
  __typename: "SiteInfo";
  /**
   * Whether or not face detection is enabled and working
   */
  faceDetectionEnabled: boolean;
  /**
   * Whether or not the map feature is enabled
   */
  mapEnabled: boolean;
}

export interface siteInfoFeatureFlags {
  siteInfo: siteInfoFeatureFlags_siteInfo;
}
