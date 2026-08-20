import { gql, useLazyQuery, useQuery } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'
import AlbumTreeNode from './AlbumTreeNode'
import {
  albumTreeActivePathQuery,
  albumTreeActivePathQueryVariables,
} from './__generated__/albumTreeActivePathQuery'
import { albumTreeRootQuery } from './__generated__/albumTreeRootQuery'

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

const AlbumTree = () => {
  const { t } = useTranslation()
  const { id: activeAlbumId } = useParams()
  const { data, loading } = useQuery<albumTreeRootQuery>(ALBUM_TREE_ROOT_QUERY)
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})

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
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }))
  }

  const roots = data?.myAlbums

  return (
    <nav
      aria-label={t('album_tree.label', 'Album tree')}
      className="overflow-y-auto h-full py-2 px-2"
    >
      {loading && !roots && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('general.loading.default', 'Loading...')}
        </div>
      )}
      {roots && roots.length === 0 && (
        <div className="px-2 py-2 text-sm text-gray-400">
          {t('album_tree.empty', 'No albums yet')}
        </div>
      )}
      <ul>
        {roots?.map(album => (
          <AlbumTreeNode
            key={album.id}
            album={album}
            depth={0}
            activeAlbumId={activeAlbumId}
            expanded={expanded}
            toggleExpand={toggleExpand}
          />
        ))}
      </ul>
    </nav>
  )
}

export default AlbumTree
