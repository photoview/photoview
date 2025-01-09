/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { ThumbnailFilter } from "./../../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: thumbnailMethodQuery
// ====================================================

export interface thumbnailMethodQuery_siteInfo {
  __typename: "SiteInfo";
  /**
   * The filter to use when generating thumbnails
   */
  thumbnailMethod: ThumbnailFilter;
}

export interface thumbnailMethodQuery {
  siteInfo: thumbnailMethodQuery_siteInfo;
}
