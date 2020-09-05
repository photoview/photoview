import React, { useEffect, useState } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'

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

function AlbumPage({ match }) {
  const albumId = match.params.id
  const showFavorites = match.params.subPage === 'favorites'

  const [onlyFavorites, setOnlyFavorites] = useState(showFavorites)

  const toggleFavorites = onlyFavorites => {
    setOnlyFavorites(onlyFavorites)
    if (onlyFavorites) {
      history.pushState(
        { showFavorites: onlyFavorites },
        '',
        '/album/' + albumId + '/favorites'
      )
    } else {
      history.back()
    }
  }

  useEffect(() => {
    const updateImageState = event => {
      setOnlyFavorites(event.state.showFavorites)
    }

    window.addEventListener('popstate', updateImageState)

    return () => {
      window.removeEventListener('popstate', updateImageState)
    }
  }, [setOnlyFavorites])

  return (
    <Query query={albumQuery} variables={{ id: albumId, onlyFavorites }}>
      {({ loading, error, data }) => {
        if (error) return <div>Error</div>

        return (
          <AlbumGallery
            album={data && data.album}
            loading={loading}
            showFavoritesToggle
            setOnlyFavorites={toggleFavorites}
            onlyFavorites={onlyFavorites}
          />
        )
      }}
    </Query>
  )
}

console.log(ReactRouterPropTypes)

AlbumPage.propTypes = {
  ...ReactRouterPropTypes,
}

export default AlbumPage
