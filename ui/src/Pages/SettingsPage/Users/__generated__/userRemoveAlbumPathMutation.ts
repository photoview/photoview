/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: userRemoveAlbumPathMutation
// ====================================================

export interface userRemoveAlbumPathMutation_userRemoveRootAlbum {
  __typename: "Album";
  id: string;
}

export interface userRemoveAlbumPathMutation {
  userRemoveRootAlbum: userRemoveAlbumPathMutation_userRemoveRootAlbum | null;
}

export interface userRemoveAlbumPathMutationVariables {
  userId: string;
  albumId: string;
}
