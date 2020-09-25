import React, { useEffect, useState } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'
import PropTypes from 'prop-types'

const albumQuery = gql`
  query albumQuery($id: Int!, $onlyFavorites: Boolean) {
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
        filter: { order_by: "title", order_direction: DESC }
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

  const toggleFavorites = refetch => {
    const newState = !onlyFavorites
    if (
      (refetchNeededAll && !newState) ||
      (refetchNeededFavorites && newState)
    ) {
      refetch({ id: albumId, onlyFavorites: newState }).then(() => {
        if (onlyFavorites) {
          refetchNeededFavorites = false
        } else {
          refetchNeededAll = false
        }
        setOnlyFavorites(newState)
      })
    } else {
      setOnlyFavorites(newState)
    }
    history.replaceState(
      {},
      '',
      '/album/' + albumId + (onlyFavorites ? '/favorites' : '')
    )
  }

  return (
    <Query query={albumQuery} variables={{ id: albumId, onlyFavorites }}>
      {({ loading, error, data, refetch }) => {
        if (error) return <div>Error</div>
        return (
          <AlbumGallery
            album={data && data.album}
            loading={loading}
            showFavoritesToggle
            setOnlyFavorites={() => {
              toggleFavorites(refetch)
            }}
            onlyFavorites={onlyFavorites}
            onFavorite={() =>
              (refetchNeededAll = refetchNeededFavorites = true)
            }
          />
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
