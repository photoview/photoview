import React, { Component } from 'react'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'
import PhotoSidebar from '../../components/sidebar/PhotoSidebar'

const photoQuery = gql`
  query allPhotosPage {
    myAlbums(orderBy: title_asc) {
      title
      id
      photos(orderBy: title_desc, first: 12) {
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
      activeAlbumIndex: -1,
      activePhotoIndex: -1,
      presenting: false,
    }

    this.setPresenting = this.setPresenting.bind(this)
    this.nextImage = this.nextImage.bind(this)
    this.previousImage = this.previousImage.bind(this)

    this.albums = []
  }

  setActiveImage(album, photo) {
    this.setState({
      activePhotoIndex: photo,
      activeAlbumIndex: album,
    })
  }

  setPresenting(presenting, index) {
    if (presenting) {
      this.setState({
        presenting: index,
      })
    } else {
      this.setState({
        presenting: false,
      })
    }
  }

  nextImage() {
    const albumImageCount = this.albums[this.state.activeAlbumIndex].photos
      .length

    if (this.state.activePhotoIndex + 1 < albumImageCount) {
      this.setState({
        activePhotoIndex: this.state.activePhotoIndex + 1,
      })
    }
  }

  previousImage() {
    if (this.state.activePhotoIndex > 0) {
      this.setState({
        activePhotoIndex: this.state.activePhotoIndex - 1,
      })
    }
  }

  render() {
    return (
      <Layout>
        <Query query={photoQuery}>
          {({ loading, error, data }) => {
            if (error) return error

            let galleryGroups = []

            this.albums = data.myAlbums

            if (data.myAlbums) {
              galleryGroups = data.myAlbums.map((album, index) => (
                <div key={album.id}>
                  <h1>{album.title}</h1>
                  <PhotoGallery
                    onSelectImage={photoIndex => {
                      this.setActiveImage(index, photoIndex)
                    }}
                    activeIndex={
                      this.state.activeAlbumIndex == index
                        ? this.state.activePhotoIndex
                        : -1
                    }
                    presenting={this.state.presenting === index}
                    setPresenting={presenting =>
                      this.setPresenting(presenting, index)
                    }
                    loading={loading}
                    photos={album.photos}
                    nextImage={this.nextImage}
                    previousImage={this.previousImage}
                  />
                </div>
              ))
            }

            let activeImage = null
            if (this.state.activeAlbumIndex != -1) {
              activeImage =
                data.myAlbums[this.state.activeAlbumIndex].photos[
                  this.state.activePhotoIndex
                ].id
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
