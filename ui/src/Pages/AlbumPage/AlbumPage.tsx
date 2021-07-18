import React, { useCallback, useEffect } from 'react'
import { useQuery, gql } from '@apollo/client'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import Layout from '../../components/layout/Layout'
import useURLParameters from '../../hooks/useURLParameters'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../../components/PaginateLoader'
import LazyLoad from '../../helpers/LazyLoad'
import { useTranslation } from 'react-i18next'
import { albumQuery, albumQueryVariables } from './__generated__/albumQuery'
import useOrderingParams from '../../hooks/useOrderingParams'

const ALBUM_QUERY = gql`
  query albumQuery(
    $id: ID!
    $onlyFavorites: Boolean
    $mediaOrderBy: String
    $mediaOrderDirection: OrderDirection
    $limit: Int
    $offset: Int
  ) {
    album(id: $id) {
      id
      title
      subAlbums(order: { order_by: "title" }) {
        id
        title
        thumbnail {
          thumbnail {
            url
          }
        }
      }
      media(
        paginate: { limit: $limit, offset: $offset }
        order: {
          order_by: $mediaOrderBy
          order_direction: $mediaOrderDirection
        }
        onlyFavorites: $onlyFavorites
      ) {
        id
        type
        thumbnail {
          url
          width
          height
        }
        highRes {
          url
        }
        videoWeb {
          url
        }
        favorite
      }
    }
  }
`

let refetchNeededAll = false
let refetchNeededFavorites = false

type AlbumPageProps = {
  match: {
    params: {
      id: string
      subPage: string
    }
  }
}

function AlbumPage({ match }: AlbumPageProps) {
  const albumId = match.params.id

  const { t } = useTranslation()

  const urlParams = useURLParameters()
  const orderParams = useOrderingParams(urlParams)

  const onlyFavorites = urlParams.getParam('favorites') == '1' ? true : false
  const setOnlyFavorites = (favorites: boolean) =>
    urlParams.setParam('favorites', favorites ? '1' : '0')

  const { loading, error, data, refetch, fetchMore } = useQuery<
    albumQuery,
    albumQueryVariables
  >(ALBUM_QUERY, {
    variables: {
      id: albumId,
      onlyFavorites,
      mediaOrderBy: orderParams.orderBy,
      mediaOrderDirection: orderParams.orderDirection,
      offset: 0,
      limit: 200,
    },
  })

  const { containerElem, finished: finishedLoadingMore } =
    useScrollPagination<albumQuery>({
      loading,
      fetchMore,
      data,
      getItems: data => data.album.media,
    })

  const toggleFavorites = useCallback(
    onlyFavorites => {
      if (
        (refetchNeededAll && !onlyFavorites) ||
        (refetchNeededFavorites && onlyFavorites)
      ) {
        refetch({ id: albumId, onlyFavorites: onlyFavorites }).then(() => {
          if (onlyFavorites) {
            refetchNeededFavorites = false
          } else {
            refetchNeededAll = false
          }
          setOnlyFavorites(onlyFavorites)
        })
      } else {
        setOnlyFavorites(onlyFavorites)
      }
    },
    [setOnlyFavorites, refetch]
  )

  useEffect(() => {
    LazyLoad.loadImages(document.querySelectorAll('img[data-src]'))
    return () => LazyLoad.disconnect()
  }, [])

  useEffect(() => {
    if (!loading) {
      LazyLoad.loadImages(document.querySelectorAll('img[data-src]'))
    }
  }, [finishedLoadingMore, onlyFavorites, loading])

  if (error) return <div>Error</div>

  return (
    <Layout
      title={
        data ? data.album.title : t('title.loading_album', 'Loading album')
      }
    >
      <AlbumGallery
        ref={containerElem}
        album={data && data.album}
        loading={loading}
        setOnlyFavorites={toggleFavorites}
        onlyFavorites={onlyFavorites}
        onFavorite={() => (refetchNeededAll = refetchNeededFavorites = true)}
        showFilter
        setOrdering={orderParams.setOrdering}
        ordering={orderParams}
      />
      <PaginateLoader
        active={!finishedLoadingMore && !loading}
        text={t('general.loading.paginate.media', 'Loading more media')}
      />
    </Layout>
  )
}

export default AlbumPage
