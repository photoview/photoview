/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: recognizeUnlabeledFaces
// ====================================================

export interface recognizeUnlabeledFaces_recognizeUnlabeledFaces {
  __typename: 'ImageFace'
  id: string
}

export interface recognizeUnlabeledFaces {
  /**
   * Check all unlabeled faces to see if they match a labeled FaceGroup, and move them if they match
   */
  recognizeUnlabeledFaces: recognizeUnlabeledFaces_recognizeUnlabeledFaces[]
}
