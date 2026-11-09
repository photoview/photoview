/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: cancelAllScanJobsMutation
// ====================================================

export interface cancelAllScanJobsMutation {
  /**
   * Cancel every scanner job currently queued or running, following the same
   * per-job semantics as cancelScanJob. Admins cancel the entire queue; other
   * users only cancel jobs for albums they own. Returns the number of jobs
   * cancelled.
   */
  cancelAllScanJobs: number;
}
