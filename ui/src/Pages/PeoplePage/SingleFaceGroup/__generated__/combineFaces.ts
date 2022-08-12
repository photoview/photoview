/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: combineFaces
// ====================================================

export interface combineFaces_combineFaceGroups {
  __typename: 'FaceGroup'
  id: string
}

export interface combineFaces {
  /**
   * Merge two face groups into a single one, all ImageFaces from source will be moved to destination
   */
  combineFaceGroups: combineFaces_combineFaceGroups
}

export interface combineFacesVariables {
  destID: string
  srcID: string
}
