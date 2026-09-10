/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: cancelScanJobMutation
// ====================================================

export interface cancelScanJobMutation {
  /**
   * Cancel a specific album's scanner job, whether queued or running. A
   * queued job is removed immediately; a running job stops once it
   * finishes its current file - already-processed files are kept, not
   * rolled back. Caller must be an admin, or hold at least UPLOAD-level
   * access on the album (the same requirement as triggering the scan).
   */
  cancelScanJob: boolean;
}

export interface cancelScanJobMutationVariables {
  albumId: string;
}
