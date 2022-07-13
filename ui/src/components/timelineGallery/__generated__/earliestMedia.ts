/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: earliestMedia
// ====================================================

export interface earliestMedia_myMedia {
  __typename: 'Media'
  id: string
  /**
   * The date the image was shot or the date it was imported as a fallback
   */
  date: Time
}

export interface earliestMedia {
  /**
   * List of media owned by the logged in user
   */
  myMedia: earliestMedia_myMedia[]
}
