import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import Layout from '../../Layout'
import PhotoSidebar from '../../components/sidebar/PhotoSidebar'
import PhotoGallery, {
  presentIndexFromHash,
} from '../../components/photoGallery/PhotoGallery'
import AlbumGallery from '../AllAlbumsPage/AlbumGallery'

const albumQuery = gql`
  query albumQuery($id: ID) {
    album(id: $id) {
      title
      subAlbums(orderBy: title_asc) {
        id
        title
        photos {
          thumbnail {
            url
          }
        }
      }
      photos(orderBy: title_desc) {
        id
        thumbnail {
          url
          width
          height
        }
      }
    }
  }
`

class AlbumPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      activeImage: -1,
      presenting: false,
    }

    const presentIndex = presentIndexFromHash(document.location.hash)
    if (presentIndex) {
      this.state.activeImage = presentIndex
      this.state.presenting = true
    }

    this.setActiveImage = this.setActiveImage.bind(this)
    this.nextImage = this.nextImage.bind(this)
    this.previousImage = this.previousImage.bind(this)
    this.setPresenting = this.setPresenting.bind(this)

    this.photos = []
  }

  setActiveImage(index) {
    this.setState({
      activeImage: index,
    })
  }

  nextImage() {
    this.setState({
      activeImage: (this.state.activeImage + 1) % this.photos.length,
    })
  }

  previousImage() {
    if (this.state.activeImage <= 0) {
      this.setState({
        activeImage: this.photos.length - 1,
      })
    } else {
      this.setState({
        activeImage: this.state.activeImage - 1,
      })
    }
  }

  componentDidUpdate(prevProps, prevState) {
    if (this.state.presenting) {
      document.history.replaceState(
        null,
        null,
        document.location.pathname + '#' + `present=${this.state.activeImage}`
      )
    } else if (presentIndexFromHash(document.location.hash)) {
      document.history.replaceState(
        null,
        null,
        document.location.pathname.split('#')[0]
      )
    }
  }

  setPresenting(presenting) {
    console.log('Presenting', presenting, this)
    this.setState({
      presenting,
    })
  }

  render() {
    const albumId = this.props.match.params.id

    return (
      <Layout>
        <Query query={albumQuery} variables={{ id: albumId }}>
          {({ loading, error, data }) => {
            if (error) return <div>Error</div>

            let subAlbumElement = null

            if (data.album) {
              this.photos = data.album.photos

              if (data.album.subAlbums.length > 0) {
                subAlbumElement = (
                  <AlbumGallery
                    loading={loading}
                    error={error}
                    albums={data.album.subAlbums}
                  />
                )
              }
            }

            return (
              <div>
                <h1>{data.album && data.album.title}</h1>
                {subAlbumElement}
                {data.album && data.album.subAlbums.length > 0 && (
                  <h2>Images</h2>
                )}
                <PhotoGallery
                  loading={loading}
                  photos={data.album && data.album.photos}
                  activeIndex={this.state.activeImage}
                  presenting={this.state.presenting}
                  onSelectImage={index => {
                    this.setActiveImage(index)
                  }}
                  setPresenting={this.setPresenting}
                  nextImage={this.nextImage}
                  previousImage={this.previousImage}
                />
                <PhotoSidebar
                  imageId={
                    this.photos.length > 0 && this.state.activeImage != -1
                      ? this.photos[this.state.activeImage].id
                      : null
                  }
                />
              </div>
            )
          }}
        </Query>
      </Layout>
    )
  }
}

export default AlbumPage
