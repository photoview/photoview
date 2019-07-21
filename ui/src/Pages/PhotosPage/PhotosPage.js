import React, { Component } from 'react'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import PhotoGallery from '../../PhotoGallery'

const photoQuery = gql`
  query allPhotosPage {
    myPhotos {
      id
      title
      thumbnail {
        url
        width
        height
      }
    }
  }
`

class PhotosPage extends Component {
  render() {
    return (
      <Layout>
        <Query query={photoQuery}>
          {({ loading, error, data }) => {
            if (error) return error

            return (
              <PhotoGallery
                loading={loading}
                title="All photos"
                photos={data.myPhotos}
              />
            )
          }}
        </Query>
      </Layout>
    )
  }
}

export default PhotosPage
