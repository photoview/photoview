import React, { useCallback, useState } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import PropTypes from 'prop-types'
import Layout from '../../Layout'

const albumQuery = gql`
  query albumQuery(
    $id: Int!
    $onlyFavorites: Boolean
    $mediaOrderBy: String
    $mediaOrderDirection: OrderDirection
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
  const [onlyFavorites, setOnlyFavorites] = useState(
    match.params.subPage === 'favorites'
  )

  const [ordering, setOrdering] = useState({
    orderBy: 'date_shot',
    orderDirection: 'ASC',
  })

  const setOrderingCallback = useCallback(
    ordering => {
      setOrdering(prevState => {
        return {
          ...prevState,
          ...ordering,
        }
      })
    },
    [setOrdering]
  )

  const toggleFavorites = useCallback(
    (onlyFavorites, refetch) => {
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
      history.replaceState(
        {},
        '',
        '/album/' + albumId + (onlyFavorites ? '/favorites' : '')
      )
    },
    [setOnlyFavorites]
  )

  return (
    <Query
      query={albumQuery}
      variables={{
        id: albumId,
        onlyFavorites,
        mediaOrderBy: ordering.orderBy,
        mediaOrderDirection: ordering.orderDirection,
      }}
    >
      {({ loading, error, data, refetch }) => {
        const setOnlyFavorites = useCallback(
          checked => {
            toggleFavorites(checked, refetch)
          },
          [toggleFavorites, refetch]
        )

        if (error) return <div>Error</div>
        return (
          <Layout title={data ? data.album.title : 'Loading album'}>
            <AlbumGallery
              album={data && data.album}
              loading={loading}
              showFavoritesToggle
              setOnlyFavorites={setOnlyFavorites}
              onlyFavorites={onlyFavorites}
              onFavorite={() =>
                (refetchNeededAll = refetchNeededFavorites = true)
              }
              showFilter
              setOrdering={setOrderingCallback}
            />
          </Layout>
        )
      }}
    </Query>
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
