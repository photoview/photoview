/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: changeScanIntervalMutation
// ====================================================

export interface changeScanIntervalMutation {
  /**
   * Set how often, in seconds, the server should automatically scan for new media,
   * a value of 0 will disable periodic scans
   */
  setPeriodicScanInterval: number
}

export interface changeScanIntervalMutationVariables {
  interval: number
}
