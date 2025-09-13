import React from 'react'
import { useTranslation } from 'react-i18next'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../components/layout/Layout'
import { useQuery, gql } from '@apollo/client'
import { getMyAlbums, getMyAlbumsVariables } from './__generated__/getMyAlbums'
import useURLParameters from '../../hooks/useURLParameters'
import useOrderingParams from '../../hooks/useOrderingParams'
import AlbumFilter from '../../components/album/AlbumFilter'

const getAlbumsQuery = gql`
  query getMyAlbums($orderBy: String, $orderDirection: OrderDirection) {
    myAlbums(
      order: { order_by: $orderBy, order_direction: $orderDirection }
      onlyRoot: true
      showEmpty: true
    ) {
      id
      title
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
  const orderParams = useOrderingParams(urlParams, 'title')

  const { error, data } = useQuery<getMyAlbums, getMyAlbumsVariables>(
    getAlbumsQuery,
    {
      variables: {
        orderBy: orderParams.orderBy,
        orderDirection: orderParams.orderDirection,
      },
    }
  )

  const sortingOptions = React.useMemo(
    () => [
      {
        value: 'updated_at',
        label: t('album_filter.sorting_options.date_imported', 'Date imported'),
      },
      {
        value: 'title',
        label: t('album_filter.sorting_options.title', 'Title'),
      },
    ],
    [t]
  )

  return (
    <Layout title="Albums">
      <AlbumFilter
        onlyFavorites={false}
        ordering={orderParams}
        setOrdering={orderParams.setOrdering}
        sortingOptions={sortingOptions}
      />
      <AlbumBoxes error={error} albums={data?.myAlbums} />
    </Layout>
  )
}

export default AlbumsPage
