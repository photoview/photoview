import React, { Component } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'

const albumQuery = gql`
  query albumQuery($id: Int!) {
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
      media(filter: { order_by: "title", order_direction: DESC }) {
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

  return (
    <Query query={albumQuery} variables={{ id: albumId }}>
      {({ loading, error, data }) => {
        if (error) return <div>Error</div>

        return <AlbumGallery album={data && data.album} loading={loading} />
      }}
    </Query>
  )
}

AlbumPage.propTypes = {
  ...ReactRouterPropTypes,
}

export default AlbumPage
