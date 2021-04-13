/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: detachImageFaces
// ====================================================

export interface detachImageFaces_detachImageFaces {
  __typename: "FaceGroup";
  id: string;
  label: string | null;
}

export interface detachImageFaces {
  /**
   * Move a list of ImageFaces to a new face group
   */
  detachImageFaces: detachImageFaces_detachImageFaces;
}

export interface detachImageFacesVariables {
  faceIDs: string[];
}
