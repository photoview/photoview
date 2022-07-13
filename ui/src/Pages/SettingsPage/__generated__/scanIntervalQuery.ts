/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: scanIntervalQuery
// ====================================================

export interface scanIntervalQuery_siteInfo {
  __typename: 'SiteInfo'
  /**
   * How often automatic scans should be initiated in seconds
   */
  periodicScanInterval: number
}

export interface scanIntervalQuery {
  siteInfo: scanIntervalQuery_siteInfo
}
