import React from 'react'
import { useTranslation } from 'react-i18next'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import { useQuery, gql } from '@apollo/client'
import { getMyAlbums, getMyAlbumsVariables } from './__generated__/getMyAlbums'
import useURLParameters from '../../hooks/useURLParameters'
import useOrderingParams from '../../hooks/useOrderingParams'
import AlbumFilter from '../../components/album/AlbumFilter'
import MobileAlbumTreeButton from '../../components/albumTree/MobileAlbumTreeButton'
import useShowHiddenAlbums from '../../hooks/useShowHiddenAlbums'

const getAlbumsQuery = gql`
  query getMyAlbums(
    $orderBy: String
    $orderDirection: OrderDirection
    $showHidden: Boolean
  ) {
    myAlbums(
      order: { order_by: $orderBy, order_direction: $orderDirection }
      onlyRoot: true
      showEmpty: true
      showHidden: $showHidden
    ) {
      id
      title
      viewerHidden
      viewerIsOwner
      thumbnail {
        id
        thumbnail {
          url
        }
      }
    }
  }
`

const AlbumsPage = () => {
  const { t } = useTranslation()

  const urlParams = useURLParameters()
  const orderParams = useOrderingParams(urlParams, 'updated_at')
  const showHidden = useShowHiddenAlbums()

  const { error, data } = useQuery<getMyAlbums, getMyAlbumsVariables>(
    getAlbumsQuery,
    {
      variables: {
        orderBy: orderParams.orderBy,
        orderDirection: orderParams.orderDirection,
        showHidden,
      },
    }
  )

  const myVolumes = data?.myAlbums.filter(a => a.viewerIsOwner)
  const sharedWithMe = data?.myAlbums.filter(a => !a.viewerIsOwner)

  const sortingOptions = React.useMemo(
    () => [
      {
        value: 'updated_at' as const,
        label: t('album_filter.sorting_options.date_imported', 'Date imported'),
      },
      {
        value: 'title' as const,
        label: t('album_filter.sorting_options.title', 'Title'),
      },
    ],
    [t]
  )

  return (
    <Layout title="Albums">
      <div className="flex items-end justify-between gap-4 flex-wrap">
        <AlbumFilter
          onlyFavorites={false}
          ordering={orderParams}
          setOrdering={orderParams.setOrdering}
          sortingOptions={sortingOptions}
        />
        <MobileAlbumTreeButton />
      </div>
      <AlbumBoxes
        error={error}
        albums={myVolumes}
        refetchQueries={['getMyAlbums']}
      />
      {sharedWithMe && sharedWithMe.length > 0 && (
        <>
          <h2 className="text-lg font-semibold mt-6 mb-2">
            {t('albums_page.shared_with_me', 'Shared with me')}
          </h2>
          <AlbumBoxes albums={sharedWithMe} refetchQueries={['getMyAlbums']} />
        </>
      )}
    </Layout>
  )
}

export default AlbumsPage
