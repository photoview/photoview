/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: userRemoveAlbumPathMutation
// ====================================================

export interface userRemoveAlbumPathMutation_userRemoveRootAlbum {
  __typename: 'Album'
  id: string
}

export interface userRemoveAlbumPathMutation {
  /**
   * Remove a root path from a user, specified by the id of the user and the top album representing the root path.
   * This album was returned when creating the path using `userAddRootPath`.
   * A list of root paths for a particular user can be retrived from the `User.rootAlbums` path.
   */
  userRemoveRootAlbum: userRemoveAlbumPathMutation_userRemoveRootAlbum | null
}

export interface userRemoveAlbumPathMutationVariables {
  userId: string
  albumId: string
}
