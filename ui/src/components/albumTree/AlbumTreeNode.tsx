import { gql, useLazyQuery } from '@apollo/client'
import React, { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { tailwindClassNames } from '../../helpers/utils'
import useShowHiddenAlbums from '../../hooks/useShowHiddenAlbums'
import {
  useHideAlbumMutation,
  toggleAlbumHidden,
} from '../albumGallery/albumHideMutations'
import {
  albumTreeSubAlbumsQuery,
  albumTreeSubAlbumsQueryVariables,
} from './__generated__/albumTreeSubAlbumsQuery'

export const ALBUM_TREE_SUB_ALBUMS_QUERY = gql`
  query albumTreeSubAlbumsQuery($id: ID!, $showHidden: Boolean) {
    album(id: $id) {
      id
      subAlbums(
        order: { order_by: "title", order_direction: ASC }
        showHidden: $showHidden
      ) {
        id
        title
        viewerHidden
      }
    }
  }
`

export type AlbumTreeNodeAlbum = {
  id: string
  title: string
  viewerHidden?: boolean
}

type AlbumTreeNodeProps = {
  album: AlbumTreeNodeAlbum
  depth: number
  activeAlbumId?: string
  expanded: Record<string, boolean>
  toggleExpand: (id: string) => void
  // When set, the tree is filtered to a search: only render descendants whose
  // id is in `visibleIds` (matched albums and their ancestors), and treat
  // every rendered node as expanded so matches are immediately visible.
  visibleIds?: Set<string>
  // Albums that directly matched the search query (as opposed to being an
  // ancestor of a match), used to highlight the actual hits.
  matchedIds?: Set<string>
  // The scrollable tree container, used to scroll a freshly expanded node's
  // children into view.
  scrollContainerRef: React.RefObject<HTMLElement>
  // Id of the album the user just expanded (not auto-expanded), so only that
  // node scrolls its children into view.
  justExpandedId: string | null
  // Reports this node's own list item element up to its parent, so the
  // parent can scroll it into view when it's the last child of an expansion.
  onNodeRef?: (id: string, el: HTMLLIElement | null) => void
  // Direct children for every visible node, fetched once up front by
  // AlbumTree in a single batched request while filtering - only set (and
  // only consulted) when visibleIds is set, so a broad match doesn't fire
  // one subAlbums request per node.
  childrenByParentId?: Map<string, AlbumTreeNodeAlbum[]>
}

const AlbumTreeNode = ({
  album,
  depth,
  activeAlbumId,
  expanded,
  toggleExpand,
  visibleIds,
  matchedIds,
  scrollContainerRef,
  justExpandedId,
  onNodeRef,
  childrenByParentId,
}: AlbumTreeNodeProps) => {
  const { t } = useTranslation()
  const isFiltering = visibleIds != null
  const isExpanded = isFiltering ? true : !!expanded[album.id]
  const isActive = activeAlbumId === album.id
  const isMatch = !!matchedIds?.has(album.id)
  const ownRef = useRef<HTMLLIElement | null>(null)
  const childRefs = useRef<Record<string, HTMLLIElement | null>>({})

  const showHidden = useShowHiddenAlbums()
  const [hideAlbum] = useHideAlbumMutation([
    'albumTreeSubAlbumsQuery',
    'albumTreeRootQuery',
    'albumTreeChildrenQuery',
  ])

  const [fetchSubAlbums, { data, loading, called, error }] = useLazyQuery<
    albumTreeSubAlbumsQuery,
    albumTreeSubAlbumsQueryVariables
  >(ALBUM_TREE_SUB_ALBUMS_QUERY, {
    variables: { id: album.id, showHidden },
  })

  useEffect(() => {
    if (isFiltering) return
    if (isExpanded && !called) {
      fetchSubAlbums()
    }
  }, [isFiltering, isExpanded, called, fetchSubAlbums])

  const subAlbums = isFiltering
    ? childrenByParentId?.get(album.id)?.filter(sub => visibleIds?.has(sub.id))
    : data?.album.subAlbums
  const hasNoChildren =
    called && !loading && !error && (subAlbums?.length ?? 0) === 0

  useEffect(() => {
    if (isFiltering || album.id !== justExpandedId) return
    if (!isExpanded || !subAlbums || subAlbums.length === 0) return

    const containerEl = scrollContainerRef.current
    const ownEl = ownRef.current
    const lastEl = childRefs.current[subAlbums[subAlbums.length - 1].id]
    if (!containerEl || !ownEl || !lastEl) return

    const containerRect = containerEl.getBoundingClientRect()
    const lastRect = lastEl.getBoundingClientRect()
    const ownRect = ownEl.getBoundingClientRect()

    const overflowBelow = lastRect.bottom - containerRect.bottom
    if (overflowBelow <= 0) return

    const ownTopAfterScroll = ownRect.top - overflowBelow
    if (ownTopAfterScroll < containerRect.top) {
      // Scrolling far enough to reveal the last child would push the
      // expanded node itself out of view, so pin that node to the top
      // instead of fully revealing the children.
      containerEl.scrollTop += ownRect.top - containerRect.top
    } else {
      containerEl.scrollTop += overflowBelow
    }
  }, [
    isFiltering,
    album.id,
    justExpandedId,
    isExpanded,
    subAlbums,
    scrollContainerRef,
  ])

  return (
    <li
      ref={el => {
        ownRef.current = el
        onNodeRef?.(album.id, el)
      }}
    >
      <div
        className="flex items-center rounded hover:bg-gray-100 dark:hover:bg-dark-bg2"
        style={{ paddingLeft: `${depth * 16}px` }}
      >
        <button
          type="button"
          aria-label={
            isExpanded
              ? t('album_tree.collapse', 'Collapse album')
              : t('album_tree.expand', 'Expand album')
          }
          onClick={() => {
            // A failed fetch leaves `called` true, so the effect above
            // won't retry it on its own - let a click while expanded and
            // errored retry directly instead of just toggling collapsed.
            if (isExpanded && error) {
              fetchSubAlbums()
            } else {
              toggleExpand(album.id)
            }
          }}
          disabled={isFiltering}
          className={tailwindClassNames(
            'w-5 h-5 flex-shrink-0 flex items-center justify-center text-gray-400 dark:text-gray-500',
            { invisible: hasNoChildren || isFiltering }
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
              'font-semibold': isMatch && !isActive,
              'opacity-50': album.viewerHidden === true,
            }
          )}
        >
          {album.title}
        </Link>
        <button
          type="button"
          title={
            album.viewerHidden
              ? t('album_tree.unhide', 'Unhide album')
              : t('album_tree.hide', 'Hide album')
          }
          aria-label={
            album.viewerHidden
              ? t('album_tree.unhide', 'Unhide album')
              : t('album_tree.hide', 'Hide album')
          }
          className="w-5 h-5 flex-shrink-0 flex items-center justify-center text-gray-400 hover:text-gray-600"
          onClick={e => {
            e.preventDefault()
            toggleAlbumHidden(hideAlbum, album.id, album.viewerHidden === true)
          }}
        >
          <span aria-hidden="true">
            {album.viewerHidden ? '\u{1F441}' : '\u{1F6AB}'}
          </span>
        </button>
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
              visibleIds={visibleIds}
              matchedIds={matchedIds}
              scrollContainerRef={scrollContainerRef}
              justExpandedId={justExpandedId}
              onNodeRef={(id, el) => {
                childRefs.current[id] = el
              }}
              childrenByParentId={childrenByParentId}
            />
          ))}
        </ul>
      )}
    </li>
  )
}

export default AlbumTreeNode
