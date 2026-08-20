import { gql, useLazyQuery } from '@apollo/client'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { tailwindClassNames } from '../../helpers/utils'
import {
  albumTreeSubAlbumsQuery,
  albumTreeSubAlbumsQueryVariables,
} from './__generated__/albumTreeSubAlbumsQuery'

export const ALBUM_TREE_SUB_ALBUMS_QUERY = gql`
  query albumTreeSubAlbumsQuery($id: ID!) {
    album(id: $id) {
      id
      subAlbums(order: { order_by: "title", order_direction: ASC }) {
        id
        title
      }
    }
  }
`

export type AlbumTreeNodeAlbum = {
  id: string
  title: string
}

type AlbumTreeNodeProps = {
  album: AlbumTreeNodeAlbum
  depth: number
  activeAlbumId?: string
  expanded: Record<string, boolean>
  toggleExpand: (id: string) => void
}

const AlbumTreeNode = ({
  album,
  depth,
  activeAlbumId,
  expanded,
  toggleExpand,
}: AlbumTreeNodeProps) => {
  const isExpanded = !!expanded[album.id]
  const isActive = activeAlbumId === album.id

  const [fetchSubAlbums, { data, loading, called }] = useLazyQuery<
    albumTreeSubAlbumsQuery,
    albumTreeSubAlbumsQueryVariables
  >(ALBUM_TREE_SUB_ALBUMS_QUERY, { variables: { id: album.id } })

  useEffect(() => {
    if (isExpanded && !called) {
      fetchSubAlbums()
    }
  }, [isExpanded, called, fetchSubAlbums])

  const subAlbums = data?.album.subAlbums
  const hasNoChildren = called && !loading && (subAlbums?.length ?? 0) === 0

  return (
    <li>
      <div
        className="flex items-center rounded hover:bg-gray-100 dark:hover:bg-dark-bg2"
        style={{ paddingLeft: `${depth * 16}px` }}
      >
        <button
          type="button"
          aria-label={isExpanded ? 'Collapse album' : 'Expand album'}
          onClick={() => toggleExpand(album.id)}
          className={tailwindClassNames(
            'w-5 h-5 flex-shrink-0 flex items-center justify-center text-gray-400 dark:text-gray-500',
            { invisible: hasNoChildren }
          )}
        >
          <svg
            viewBox="0 0 24 24"
            className={tailwindClassNames(
              'w-3 h-3 transition-transform fill-current',
              { 'rotate-90': isExpanded }
            )}
          >
            <path d="M8 5v14l11-7z" />
          </svg>
        </button>
        <Link
          to={`/album/${album.id}`}
          title={album.title}
          className={tailwindClassNames(
            'truncate text-sm py-1 px-1 rounded flex-1 min-w-0',
            {
              'font-semibold text-blue-600 dark:text-blue-400': isActive,
            }
          )}
        >
          {album.title}
        </Link>
      </div>
      {isExpanded && subAlbums && subAlbums.length > 0 && (
        <ul>
          {subAlbums.map(sub => (
            <AlbumTreeNode
              key={sub.id}
              album={sub}
              depth={depth + 1}
              activeAlbumId={activeAlbumId}
              expanded={expanded}
              toggleExpand={toggleExpand}
            />
          ))}
        </ul>
      )}
    </li>
  )
}

export default AlbumTreeNode
