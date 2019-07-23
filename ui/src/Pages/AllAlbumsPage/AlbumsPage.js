import React, { Component } from 'react'
import AlbumGallery from './AlbumGallery'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'

const getAlbumsQuery = gql`
  query getMyAlbums {
    myAlbums(filter: { parentAlbum: null }, orderBy: title_asc) {
      id
      title
      photos {
        thumbnail {
          url
        }
      }
    }
  }
`

class AlbumsPage extends Component {
  render() {
    return (
      <Layout>
        <h1>Albums</h1>
        <Query query={getAlbumsQuery}>
          {({ loading, error, data }) => (
            <AlbumGallery
              loading={loading}
              error={error}
              albums={data && data.myAlbums}
            />
          )}
        </Query>
      </Layout>
    )
  }
}

export default AlbumsPage
