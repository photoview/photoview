import React, { useEffect, useReducer, useState } from 'react'
import { useTranslation } from 'react-i18next'
import AlbumTitle from '../album/AlbumTitle'
import MediaGallery, {
  MEDIA_GALLERY_FRAGMENT,
} from '../photoGallery/MediaGallery'
import AlbumBoxes from './AlbumBoxes'
import AlbumFilter from '../album/AlbumFilter'
import {
  mediaGalleryReducer,
  urlPresentModeSetupHook,
} from '../photoGallery/mediaGalleryReducer'
import { MediaOrdering, SetOrderingFn } from '../../hooks/useOrderingParams'
import { gql, useMutation } from '@apollo/client'
import { Button } from '../../primitives/form/Input'
import Modal from '../../primitives/Modal'
import { MessageState } from '../messages/Messages'
import { NotificationType } from '../../__generated__/globalTypes'
import { AlbumGalleryFields } from './__generated__/AlbumGalleryFields'
import {
  deleteMediaList,
  deleteMediaListVariables,
} from './__generated__/deleteMediaList'

export const ALBUM_GALLERY_FRAGMENT = gql`
  ${MEDIA_GALLERY_FRAGMENT}

  fragment AlbumGalleryFields on Album {
    id
    title
    viewerCanUpload
    subAlbums(
      order: { order_by: "title", order_direction: $orderDirection }
      showHidden: $showHidden
    ) {
      id
      title
      viewerHidden
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
    media(
      paginate: { limit: $limit, offset: $offset }
      order: { order_by: $mediaOrderBy, order_direction: $orderDirection }
      onlyFavorites: $onlyFavorites
    ) {
      ...MediaGalleryFields
    }
  }
`

const DELETE_MEDIA_LIST_MUTATION = gql`
  mutation deleteMediaList($mediaIds: [ID!]!) {
    deleteMediaList(mediaIds: $mediaIds) {
      mediaId
      success
      error
    }
  }
`

type AlbumGalleryProps = {
  album?: AlbumGalleryFields
  loading?: boolean
  customAlbumLink?(albumID: string): string
  showFilter?: boolean
  setOnlyFavorites?(favorites: boolean): void
  setOrdering?: SetOrderingFn
  ordering?: MediaOrdering
  onlyFavorites?: boolean
  onFavorite?(): void
}

const AlbumGallery = React.forwardRef(
  (
    {
      album,
      loading = false,
      customAlbumLink,
      showFilter = false,
      setOnlyFavorites,
      setOrdering,
      ordering,
      onlyFavorites = false,
    }: AlbumGalleryProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    const { t } = useTranslation()
    const [mediaState, dispatchMedia] = useReducer(mediaGalleryReducer, {
      presenting: false,
      activeIndex: -1,
      media: album?.media || [],
    })

    const [selectMode, setSelectMode] = useState(false)
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
    const [showConfirmDelete, setShowConfirmDelete] = useState(false)

    const [deleteMediaList, { loading: deleting }] = useMutation<
      deleteMediaList,
      deleteMediaListVariables
    >(DELETE_MEDIA_LIST_MUTATION, {
      refetchQueries: ['albumQuery'],
      onCompleted: data => {
        setSelectMode(false)
        setSelectedIds(new Set())

        const results = data.deleteMediaList
        const failed = results.filter(r => !r.success)
        const notifyKey = `delete-media-${Date.now()}`
        MessageState.add({
          key: notifyKey,
          type: NotificationType.Message,
          props:
            failed.length === 0
              ? {
                  header: t(
                    'album_gallery.select.delete_complete',
                    'Deleted {{count}} files',
                    { count: results.length }
                  ),
                }
              : {
                  header: t(
                    'album_gallery.select.delete_finished_with_errors',
                    'Deleted {{success}} of {{total}} files',
                    {
                      success: results.length - failed.length,
                      total: results.length,
                    }
                  ),
                  content: failed.map(r => r.error).join('\n'),
                },
        })
        setTimeout(() => MessageState.removeKey(notifyKey), 4000)
      },
      // The global Apollo error link already shows a toast for network/
      // GraphQL-level failures; this only exists so Apollo treats the
      // mutation as handled instead of leaving an unhandled promise
      // rejection, and so the selection stays intact for a retry instead
      // of being silently cleared as if the delete had succeeded.
      onError: () => undefined,
    })

    useEffect(() => {
      dispatchMedia({ type: 'replaceMedia', media: album?.media || [] })
    }, [album?.media])

    urlPresentModeSetupHook({
      dispatchMedia,
      openPresentMode: event => {
        dispatchMedia({
          type: 'openPresentMode',
          activeIndex: event.state.activeIndex,
        })
      },
    })

    const toggleSelectMode = () => {
      setSelectMode(mode => !mode)
      setSelectedIds(new Set())
    }

    const toggleSelected = (mediaId: string) => {
      setSelectedIds(prev => {
        const next = new Set(prev)
        if (next.has(mediaId)) {
          next.delete(mediaId)
        } else {
          next.add(mediaId)
        }
        return next
      })
    }

    const selectAll = () => {
      setSelectedIds(new Set(mediaState.media.map(media => media.id)))
    }

    let subAlbumElement = null
    if (album) {
      if (album.subAlbums.length > 0) {
        subAlbumElement = (
          <AlbumBoxes
            albums={album.subAlbums}
            getCustomLink={customAlbumLink}
            refetchQueries={['albumQuery']}
          />
        )
      }
    } else {
      subAlbumElement = <AlbumBoxes />
    }

    return (
      <div ref={ref}>
        {showFilter && (
          <AlbumFilter
            onlyFavorites={onlyFavorites}
            setOnlyFavorites={setOnlyFavorites}
            setOrdering={setOrdering}
            ordering={ordering}
          />
        )}
        <AlbumTitle album={album} disableLink />
        {album?.viewerCanUpload && mediaState.media.length > 0 && (
          <div className="flex items-center gap-2 mb-2">
            {selectMode ? (
              <>
                <span>
                  {t('album_gallery.select.count', '{{count}} selected', {
                    count: selectedIds.size,
                  })}
                </span>
                <Button onClick={selectAll}>
                  {t('album_gallery.select.select_all', 'Select all')}
                </Button>
                <Button onClick={toggleSelectMode}>
                  {t('general.action.cancel', 'Cancel')}
                </Button>
                <Button
                  variant="negative"
                  disabled={selectedIds.size === 0 || deleting}
                  onClick={() => setShowConfirmDelete(true)}
                >
                  {t('album_gallery.select.delete_selected', 'Delete selected')}
                </Button>
              </>
            ) : (
              <Button onClick={toggleSelectMode}>
                {t('album_gallery.select.select', 'Select')}
              </Button>
            )}
          </div>
        )}
        {subAlbumElement}
        <MediaGallery
          loading={loading}
          mediaState={mediaState}
          dispatchMedia={dispatchMedia}
          selectMode={selectMode}
          selectedIds={selectedIds}
          onToggleSelect={toggleSelected}
        />
        <Modal
          open={showConfirmDelete}
          onClose={() => setShowConfirmDelete(false)}
          title={t('album_gallery.select.confirm_delete.title', 'Delete files')}
          description={t(
            'album_gallery.select.confirm_delete.description',
            'Move {{count}} files to the trash? This can be recovered by an administrator, but they will disappear from the library immediately.',
            { count: selectedIds.size }
          )}
          actions={[
            {
              key: 'cancel',
              label: t('general.action.cancel', 'Cancel'),
              onClick: () => setShowConfirmDelete(false),
            },
            {
              key: 'delete',
              label: t(
                'album_gallery.select.delete_selected',
                'Delete selected'
              ),
              variant: 'negative',
              onClick: () => {
                setShowConfirmDelete(false)
                deleteMediaList({ variables: { mediaIds: [...selectedIds] } })
              },
            },
          ]}
        />
      </div>
    )
  }
)

export default AlbumGallery
