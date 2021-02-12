import React, { useCallback } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import { useQuery, gql } from '@apollo/client'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import PropTypes from 'prop-types'
import Layout from '../../Layout'
import useURLParameters from '../../hooks/useURLParameters'
import useScrollPagination from '../../hooks/useScrollPagination'
import { Loader } from 'semantic-ui-react'

const albumQuery = gql`
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
      subAlbums(filter: { order_by: "title" }) {
        id
        title
        thumbnail {
          thumbnail {
            url
          }
        }
      }
      media(
        filter: {
          limit: $limit
          offset: $offset
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

function AlbumPage({ match }) {
  const albumId = match.params.id

  const { getParam, setParam, setParams } = useURLParameters()

  const onlyFavorites = getParam('favorites') == '1' ? true : false
  const setOnlyFavorites = favorites => setParam('favorites', favorites ? 1 : 0)

  const orderBy = getParam('orderBy', 'date_shot')
  const orderDirection = getParam('orderDirection', 'ASC')

  const setOrdering = useCallback(
    ({ orderBy, orderDirection }) => {
      let updatedParams = []
      if (orderBy !== undefined) {
        updatedParams.push({ key: 'orderBy', value: orderBy })
      }
      if (orderDirection !== undefined) {
        updatedParams.push({ key: 'orderDirection', value: orderDirection })
      }

      setParams(updatedParams)
    },
    [setParams]
  )

  const { loading, error, data, refetch, fetchMore } = useQuery(albumQuery, {
    variables: {
      id: albumId,
      onlyFavorites,
      mediaOrderBy: orderBy,
      mediaOrderDirection: orderDirection,
      offset: 0,
      limit: 200,
    },
  })

  const { containerElem, finished: finishedLoadingMore } = useScrollPagination({
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

  if (error) return <div>Error</div>

  return (
    <Layout title={data ? data.album.title : 'Loading album'}>
      <AlbumGallery
        ref={containerElem}
        album={data && data.album}
        loading={loading}
        showFavoritesToggle
        setOnlyFavorites={toggleFavorites}
        onlyFavorites={onlyFavorites}
        onFavorite={() => (refetchNeededAll = refetchNeededFavorites = true)}
        showFilter
        setOrdering={setOrdering}
        ordering={{ orderBy, orderDirection }}
      />
      <Loader
        style={{ margin: '42px 0 24px 0' }}
        active={!finishedLoadingMore && !loading}
        inline="centered"
      >
        Loading more media
      </Loader>
    </Layout>
  )
}

AlbumPage.propTypes = {
  ...ReactRouterPropTypes,
  match: PropTypes.shape({
    params: PropTypes.shape({
      id: PropTypes.string,
      subPage: PropTypes.string,
    }),
  }),
}

export default AlbumPage
