import React, { Component } from 'react'
import AlbumBoxes from '../../components/albumGallery/AlbumBoxes'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'

const getAlbumsQuery = gql`
  query getMyAlbums {
    myAlbums(filter: { order_by: "title" }, onlyRoot: true, showEmpty: true) {
      id
      title
      thumbnail {
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
      <Layout title="Albums">
        <h1>Albums</h1>
        <Query query={getAlbumsQuery}>
          {({ loading, error, data }) => (
            <AlbumBoxes
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
