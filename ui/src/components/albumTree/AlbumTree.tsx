import { gql, useLazyQuery, useQuery } from '@apollo/client'
import React, { useContext, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'
import { debounce, DebouncedFn } from '../../helpers/utils'
import AlbumTreeNode, { AlbumTreeNodeAlbum } from './AlbumTreeNode'
import { AlbumTreeSearchContext } from './AlbumTreeSearchContext'
import {
  albumTreeActivePathQuery,
  albumTreeActivePathQueryVariables,
} from './__generated__/albumTreeActivePathQuery'
import {
  albumTreeRootQuery,
  albumTreeRootQueryVariables,
  albumTreeRootQuery_myAlbums,
} from './__generated__/albumTreeRootQuery'
import {
  albumTreeSearchQuery,
  albumTreeSearchQueryVariables,
  albumTreeSearchQuery_search_albums,
} from './__generated__/albumTreeSearchQuery'
import {
  albumTreeChildrenQuery,
  albumTreeChildrenQueryVariables,
} from './__generated__/albumTreeChildrenQuery'

export const ALBUM_TREE_ROOT_QUERY = gql`
  query albumTreeRootQuery {
    myAlbums(
      onlyRoot: true
      showEmpty: true
      order: { order_by: "title", order_direction: ASC }
    ) {
      id
      title
    }
  }
`

export const ALBUM_TREE_ACTIVE_PATH_QUERY = gql`
  query albumTreeActivePathQuery($id: ID!) {
    album(id: $id) {
      id
      path {
        id
      }
    }
  }
`

// The filter expands every match together with its ancestors, so both halves
// of that work have to stay bounded. The expensive half is the ancestor path,
// which the server resolves per matched album - on a library of a few thousand
// albums an unbounded filter spends most of a second there alone. The node
// limit then keeps the batch that follows below the server's own cap, which a
// deep tree could otherwise cross and lose the whole filter view to an error.
export const TREE_FILTER_MATCH_LIMIT = 100
export const TREE_FILTER_NODE_LIMIT = 400

export const ALBUM_TREE_SEARCH_QUERY = gql`
  query albumTreeSearchQuery($query: String!, $limitAlbums: Int) {
    search(query: $query, limitAlbums: $limitAlbums, limitMedia: 0) {
      albums {
        id
        path {
          id
        }
      }
    }
  }
`

export const ALBUM_TREE_CHILDREN_QUERY = gql`
  query albumTreeChildrenQuery($albumIds: [ID!]!) {
    albumTreeChildren(albumIds: $albumIds) {
      albumId
      children {
        id
        title
      }
    }
  }
`

/**
 * Picks which nodes the filtered tree shows: every match it keeps, together
 * with the complete chain of ancestors that leads to it.
 *
 * A match is only useful if the tree can be walked down to it from a root, so
 * a match and its path are taken or dropped as one - filling the node budget
 * with matches first would leave some of them stranded below a missing
 * ancestor, and the tree would claim there were no matches at all while the
 * search had plenty.
 */
export const selectFilteredNodes = (
  matches: albumTreeSearchQuery_search_albums[]
) => {
  const matchedIds = new Set<string>()
  let visibleIds = new Set<string>()
  let truncated = matches.length >= TREE_FILTER_MATCH_LIMIT

  for (const album of matches) {
    const candidate = new Set(visibleIds)
    candidate.add(album.id)
    for (const ancestor of album.path) {
      candidate.add(ancestor.id)
    }

    if (candidate.size > TREE_FILTER_NODE_LIMIT) {
      truncated = true

      continue
    }

    matchedIds.add(album.id)
    visibleIds = candidate
  }

  return { matchedIds, visibleIds, truncated }
}

const AlbumTree = () => {
  const { t } = useTranslation()
  const { id: activeAlbumId } = useParams()
  const { data, loading, error } = useQuery<
    albumTreeRootQuery,
    albumTreeRootQueryVariables
  >(ALBUM_TREE_ROOT_QUERY)
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})
  const [justExpandedId, setJustExpandedId] = useState<string | null>(null)
  const scrollContainerRef = useRef<HTMLElement>(null)

  const [fetchActivePath, { data: activePathData }] = useLazyQuery<
    albumTreeActivePathQuery,
    albumTreeActivePathQueryVariables
  >(ALBUM_TREE_ACTIVE_PATH_QUERY)

  useEffect(() => {
    if (activeAlbumId) {
      fetchActivePath({ variables: { id: activeAlbumId } })
    }
  }, [activeAlbumId, fetchActivePath])

  useEffect(() => {
    if (!activeAlbumId) return

    const ancestorIds = activePathData?.album.path.map(a => a.id) ?? []

    setExpanded(prev => {
      const next = { ...prev }
      for (const id of [...ancestorIds, activeAlbumId]) {
        next[id] = true
      }
      return next
    })
  }, [activeAlbumId, activePathData])

  const toggleExpand = (id: string) => {
    const nowExpanded = !expanded[id]
    setExpanded(prev => ({ ...prev, [id]: nowExpanded }))
    setJustExpandedId(nowExpanded ? id : null)
  }

  const { query: searchQuery } = useContext(AlbumTreeSearchContext)
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('')

  const debouncedSetQuery = useRef<null | DebouncedFn<(query: string) => void>>(
    null
  )
  useEffect(() => {
    debouncedSetQuery.current = debounce<(query: string) => void>(
      query => setDebouncedSearchQuery(query),
      250
    )
    return () => debouncedSetQuery.current?.cancel()
  }, [])

  useEffect(() => {
    if (searchQuery.trim() === '') {
      // A call scheduled before the field was cleared would otherwise still
      // fire and restore the stale query, leaving the tree filtered while
      // the search field itself is empty.
      debouncedSetQuery.current?.cancel()
      setDebouncedSearchQuery('')
      return
    }
    debouncedSetQuery.current?.(searchQuery.trim())
  }, [searchQuery])

  const [
    fetchTreeSearch,
    {
      data: treeSearchData,
      loading: treeSearchLoading,
      error: treeSearchError,
    },
  ] = useLazyQuery<albumTreeSearchQuery, albumTreeSearchQueryVariables>(
    ALBUM_TREE_SEARCH_QUERY
  )

  useEffect(() => {
    if (debouncedSearchQuery !== '') {
      fetchTreeSearch({
        variables: {
          query: debouncedSearchQuery,
          limitAlbums: TREE_FILTER_MATCH_LIMIT,
        },
      })
    }
  }, [debouncedSearchQuery, fetchTreeSearch])

  const isFiltering = debouncedSearchQuery !== ''

  let matchedIds: Set<string> | undefined
  let visibleIds: Set<string> | undefined
  let filterTruncated = false
  if (isFiltering && treeSearchData) {
    const selection = selectFilteredNodes(treeSearchData.search.albums)
    matchedIds = selection.matchedIds
    visibleIds = selection.visibleIds
    filterTruncated = selection.truncated
  }

  const [
    fetchTreeChildren,
    { data: treeChildrenData, error: treeChildrenError },
  ] = useLazyQuery<albumTreeChildrenQuery, albumTreeChildrenQueryVariables>(
    ALBUM_TREE_CHILDREN_QUERY
  )

  useEffect(() => {
    if (visibleIds) {
      fetchTreeChildren({
        variables: { albumIds: Array.from(visibleIds) },
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [treeSearchData, fetchTreeChildren])

  const childrenByParentId = new Map<string, AlbumTreeNodeAlbum[]>()
  if (treeChildrenData) {
    for (const entry of treeChildrenData.albumTreeChildren) {
      childrenByParentId.set(entry.albumId, entry.children)
    }
  }

  const roots = data?.myAlbums
  const visibleRoots = isFiltering
    ? roots?.filter(album => visibleIds?.has(album.id))
    : roots

  const renderNode = (album: albumTreeRootQuery_myAlbums) => (
    <AlbumTreeNode
      key={album.id}
      album={album}
      depth={0}
      activeAlbumId={activeAlbumId}
      expanded={expanded}
      toggleExpand={toggleExpand}
      visibleIds={visibleIds}
      matchedIds={matchedIds}
      scrollContainerRef={scrollContainerRef}
      justExpandedId={justExpandedId}
      childrenByParentId={isFiltering ? childrenByParentId : undefined}
    />
  )

  return (
    <nav
      ref={scrollContainerRef}
      aria-label={t('album_tree.label', 'Album tree')}
      className="overflow-y-auto h-full py-2 px-2"
    >
      {loading && !roots && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('general.loading.default', 'Loading...')}
        </div>
      )}
      {error && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.error', 'Could not load albums')}
        </div>
      )}
      {roots && roots.length === 0 && !isFiltering && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.empty', 'No albums yet')}
        </div>
      )}
      {isFiltering && treeSearchLoading && !treeSearchData && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('general.loading.default', 'Loading...')}
        </div>
      )}
      {filterTruncated && !treeSearchError && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t(
            'album_tree.filter_truncated',
            'Showing the first {{limit}} matching albums',
            { limit: TREE_FILTER_MATCH_LIMIT }
          )}
        </div>
      )}
      {isFiltering && treeSearchError && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.search_error', 'Could not load matching albums')}
        </div>
      )}
      {isFiltering && treeChildrenError && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t(
            'album_tree.children_error',
            'Could not load albums below the matches'
          )}
        </div>
      )}
      {isFiltering && treeSearchData && visibleRoots?.length === 0 && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.no_matches', 'No matching albums')}
        </div>
      )}
      <ul>{visibleRoots?.map(renderNode)}</ul>
    </nav>
  )
}

export default AlbumTree
