/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: mapboxToken
// ====================================================

export interface mapboxToken {
  /**
   * Get the mapbox api token, returns null if mapbox is not enabled
   */
  mapboxToken: string | null;
  /**
   * Get media owned by the logged in user, returned in GeoJson format
   */
  myMediaGeoJson: Any;
}
