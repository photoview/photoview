/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { ThumbnailFilter } from './../../../__generated__/globalTypes'

// ====================================================
// GraphQL mutation operation: setThumbnailMethodMutation
// ====================================================

export interface setThumbnailMethodMutation {
  /**
   * Set the filter to be used when generating thumbnails
   */
  setThumbnailDownsampleMethod: ThumbnailFilter
}

export interface setThumbnailMethodMutationVariables {
  method: ThumbnailFilter
}
