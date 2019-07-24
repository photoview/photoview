import React, { Component } from 'react'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import PhotoGallery from '../../PhotoGallery'
import PhotoSidebar from '../../components/sidebar/PhotoSidebar'

const photoQuery = gql`
  query allPhotosPage {
    myAlbums(orderBy: title_asc) {
      title
      id
      photos(orderBy: title_desc) {
        id
        title
        thumbnail {
          url
          width
          height
        }
      }
    }
  }
`

class PhotosPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      activeAlbum: null,
      activeIndex: null,
    }
  }

  setActiveImage(album, index) {
    this.setState({
      activeIndex: index,
      activeAlbum: album,
    })
  }

  render() {
    return (
      <Layout>
        <Query query={photoQuery}>
          {({ loading, error, data }) => {
            if (error) return error

            let galleryGroups = []

            if (data.myAlbums) {
              galleryGroups = data.myAlbums.map(album => (
                <div key={album.id}>
                  <h1>{album.title}</h1>
                  <PhotoGallery
                    onSelectImage={index => {
                      this.setActiveImage(album.id, index)
                    }}
                    activeIndex={
                      this.state.activeAlbum == album.id
                        ? this.state.activeIndex
                        : -1
                    }
                    loading={loading}
                    photos={album.photos}
                  />
                </div>
              ))
            }

            let activeImage = null
            if (this.state.activeAlbum) {
              activeImage = data.myAlbums.find(
                album => album.id == this.state.activeAlbum
              ).photos[this.state.activeIndex].id
            }

            return (
              <div>
                {galleryGroups}
                <PhotoSidebar imageId={activeImage} />
              </div>
            )
          }}
        </Query>
      </Layout>
    )
  }
}

export default PhotosPage
