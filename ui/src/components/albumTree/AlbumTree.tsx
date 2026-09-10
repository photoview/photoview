import { gql, useLazyQuery, useQuery } from '@apollo/client'
import React, { useContext, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'
import { debounce, DebouncedFn } from '../../helpers/utils'
import useShowHiddenAlbums from '../../hooks/useShowHiddenAlbums'
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
} from './__generated__/albumTreeSearchQuery'
import {
  albumTreeChildrenQuery,
  albumTreeChildrenQueryVariables,
} from './__generated__/albumTreeChildrenQuery'

export const ALBUM_TREE_ROOT_QUERY = gql`
  query albumTreeRootQuery($showHidden: Boolean) {
    myAlbums(
      onlyRoot: true
      showEmpty: true
      showHidden: $showHidden
      order: { order_by: "title", order_direction: ASC }
    ) {
      id
      title
      viewerHidden
      viewerIsOwner
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

export const ALBUM_TREE_SEARCH_QUERY = gql`
  query albumTreeSearchQuery($query: String!, $showHidden: Boolean) {
    search(
      query: $query
      limitAlbums: 0
      limitMedia: 0
      showHidden: $showHidden
    ) {
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
  query albumTreeChildrenQuery($albumIds: [ID!]!, $showHidden: Boolean) {
    albumTreeChildren(albumIds: $albumIds, showHidden: $showHidden) {
      albumId
      children {
        id
        title
        viewerHidden
      }
    }
  }
`

const AlbumTree = () => {
  const { t } = useTranslation()
  const { id: activeAlbumId } = useParams()
  const showHidden = useShowHiddenAlbums()
  const { data, loading, error } = useQuery<
    albumTreeRootQuery,
    albumTreeRootQueryVariables
  >(ALBUM_TREE_ROOT_QUERY, { variables: { showHidden } })
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
        variables: { query: debouncedSearchQuery, showHidden },
      })
    }
  }, [debouncedSearchQuery, showHidden, fetchTreeSearch])

  const isFiltering = debouncedSearchQuery !== ''

  let matchedIds: Set<string> | undefined
  let visibleIds: Set<string> | undefined
  if (isFiltering && treeSearchData) {
    matchedIds = new Set(treeSearchData.search.albums.map(a => a.id))
    visibleIds = new Set(matchedIds)
    for (const album of treeSearchData.search.albums) {
      for (const ancestor of album.path) {
        visibleIds.add(ancestor.id)
      }
    }
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
        variables: { albumIds: Array.from(visibleIds), showHidden },
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [treeSearchData, showHidden, fetchTreeChildren])

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

  const myVolumes = visibleRoots?.filter(a => a.viewerIsOwner)
  const sharedWithMe = visibleRoots?.filter(a => !a.viewerIsOwner)

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
      <ul>{myVolumes?.map(renderNode)}</ul>
      {sharedWithMe && sharedWithMe.length > 0 && (
        <>
          <div className="px-2 pt-3 pb-1 text-xs font-semibold text-gray-400 uppercase">
            {t('album_tree.shared_with_me', 'Shared with me')}
          </div>
          <ul>{sharedWithMe.map(renderNode)}</ul>
        </>
      )}
    </nav>
  )
}

export default AlbumTree
