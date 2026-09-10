/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: albumTreeChildrenQuery
// ====================================================

export interface albumTreeChildrenQuery_albumTreeChildren_children {
  __typename: 'Album'
  id: string
  title: string
  /**
   * Whether the currently logged in user has personally hidden this album from their own navigation. Never affects other users.
   */
  viewerHidden: boolean
}

export interface albumTreeChildrenQuery_albumTreeChildren {
  __typename: 'AlbumTreeChildren'
  albumId: string
  children: albumTreeChildrenQuery_albumTreeChildren_children[]
}

export interface albumTreeChildrenQuery {
  /**
   * Direct children for every album in albumIds, in one round trip - used by
   * the album tree's search/filter view so a broad match doesn't fire one
   * subAlbums request per matched node and per ancestor.
   */
  albumTreeChildren: albumTreeChildrenQuery_albumTreeChildren[]
}

export interface albumTreeChildrenQueryVariables {
  albumIds: string[]
  showHidden?: boolean | null
}
