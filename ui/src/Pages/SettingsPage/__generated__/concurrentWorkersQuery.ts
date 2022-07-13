/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: concurrentWorkersQuery
// ====================================================

export interface concurrentWorkersQuery_siteInfo {
  __typename: 'SiteInfo'
  /**
   * How many max concurrent scanner jobs that should run at once
   */
  concurrentWorkers: number
}

export interface concurrentWorkersQuery {
  siteInfo: concurrentWorkersQuery_siteInfo
}
