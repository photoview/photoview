/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: deleteMediaList
// ====================================================

export interface deleteMediaList_deleteMediaList {
  __typename: "DeleteMediaResult";
  mediaId: string;
  success: boolean;
  /**
   * Present only when success is false
   */
  error: string | null;
}

export interface deleteMediaList {
  /**
   * Delete several media files at once, same rules as deleteMedia. Each
   * file is attempted independently - one failure doesn't stop the rest.
   */
  deleteMediaList: deleteMediaList_deleteMediaList[];
}

export interface deleteMediaListVariables {
  mediaIds: string[];
}
