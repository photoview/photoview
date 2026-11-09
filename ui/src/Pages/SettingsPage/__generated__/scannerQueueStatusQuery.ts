/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { ScannerJobStatus } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: scannerQueueStatusQuery
// ====================================================

export interface scannerQueueStatusQuery_scannerQueueStatus_album_path {
  __typename: "Album";
  id: string;
  title: string;
}

export interface scannerQueueStatusQuery_scannerQueueStatus_album {
  __typename: "Album";
  id: string;
  title: string;
  /**
   * A breadcrumb list of all parent albums down to this one
   */
  path: scannerQueueStatusQuery_scannerQueueStatus_album_path[];
}

export interface scannerQueueStatusQuery_scannerQueueStatus {
  __typename: "ScannerQueueItem";
  status: ScannerJobStatus;
  album: scannerQueueStatusQuery_scannerQueueStatus_album;
}

export interface scannerQueueStatusQuery {
  /**
   * Snapshot of what the scanner is currently running or has queued.
   * Admins see every job; other users only see jobs for albums they own.
   */
  scannerQueueStatus: scannerQueueStatusQuery_scannerQueueStatus[];
}
