import React from 'react'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'

const albumQuery = gql`
  query getAlbumSidebar($id: ID!) {
    album(id: $id) {
      id
      title
    }
  }
`

const AlbumSidebar = ({ albumId }) => {
  return (
    <div>
      <p>Album options</p>
      <Query query={albumQuery} variables={{ id: albumId }}>
        {({ loading, error, data }) => {
          if (loading) return <div>Loading...</div>
          if (error) return <div>{error.message}</div>

          return <h1>{data.album.title}</h1>
        }}
      </Query>
    </div>
  )
}

export default AlbumSidebar
