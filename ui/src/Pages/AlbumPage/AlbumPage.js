import React, { useCallback, useState, useEffect } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import { useLocation } from 'react-router-dom'
import { useQuery, gql } from '@apollo/client'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import PropTypes from 'prop-types'
import Layout from '../../Layout'

const albumQuery = gql`
  query albumQuery(
    $id: ID!
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

  const urlParams = new URLSearchParams(useLocation().search)
  const [ordering, setOrdering] = useState({
    orderBy: urlParams.get('orderBy') || 'date_shot',
    orderDirection: urlParams.get('orderDirection') || 'ASC',
  })

  const { loading, error, data, refetch } = useQuery(albumQuery, {
    variables: {
      id: albumId,
      onlyFavorites,
      mediaOrderBy: ordering.orderBy,
      mediaOrderDirection: ordering.orderDirection,
    },
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
    [setOrdering, onlyFavorites]
  )

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
    const pathName = `/album/${albumId + (onlyFavorites ? '/favorites' : '')}`
    const queryString = `orderBy=${ordering.orderBy}&orderDirection=${ordering.orderDirection}`

    history.replaceState({}, '', pathName + '?' + queryString)
  }, [onlyFavorites, ordering])

  if (error) return <div>Error</div>

  return (
    <Layout title={data ? data.album.title : 'Loading album'}>
      <AlbumGallery
        album={data && data.album}
        loading={loading}
        showFavoritesToggle
        setOnlyFavorites={toggleFavorites}
        onlyFavorites={onlyFavorites}
        onFavorite={() => (refetchNeededAll = refetchNeededFavorites = true)}
        showFilter
        setOrdering={setOrderingCallback}
        ordering={ordering}
      />
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
