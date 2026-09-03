import { gql, useLazyQuery, useQuery } from '@apollo/client'
import React, { useContext, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'
import { debounce, DebouncedFn } from '../../helpers/utils'
import AlbumTreeNode from './AlbumTreeNode'
import { AlbumTreeSearchContext } from './AlbumTreeSearchContext'
import {
  albumTreeActivePathQuery,
  albumTreeActivePathQueryVariables,
} from './__generated__/albumTreeActivePathQuery'
import { albumTreeRootQuery } from './__generated__/albumTreeRootQuery'
import {
  albumTreeSearchQuery,
  albumTreeSearchQueryVariables,
} from './__generated__/albumTreeSearchQuery'

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

export const ALBUM_TREE_SEARCH_QUERY = gql`
  query albumTreeSearchQuery($query: String!) {
    search(query: $query, limitAlbums: 0, limitMedia: 0) {
      albums {
        id
        path {
          id
        }
      }
    }
  }
`

const AlbumTree = () => {
  const { t } = useTranslation()
  const { id: activeAlbumId } = useParams()
  const { data, loading } = useQuery<albumTreeRootQuery>(ALBUM_TREE_ROOT_QUERY)
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
      setDebouncedSearchQuery('')
      return
    }
    debouncedSetQuery.current?.(searchQuery.trim())
  }, [searchQuery])

  const [fetchTreeSearch, { data: treeSearchData, loading: treeSearchLoading }] =
    useLazyQuery<albumTreeSearchQuery, albumTreeSearchQueryVariables>(
      ALBUM_TREE_SEARCH_QUERY
    )

  useEffect(() => {
    if (debouncedSearchQuery !== '') {
      fetchTreeSearch({ variables: { query: debouncedSearchQuery } })
    }
  }, [debouncedSearchQuery, fetchTreeSearch])

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

  const roots = data?.myAlbums
  const visibleRoots = isFiltering
    ? roots?.filter(album => visibleIds?.has(album.id))
    : roots

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
      {isFiltering && treeSearchData && visibleRoots?.length === 0 && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.no_matches', 'No matching albums')}
        </div>
      )}
      <ul>
        {visibleRoots?.map(album => (
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
          />
        ))}
      </ul>
    </nav>
  )
}

export default AlbumTree
