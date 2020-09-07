import React, { Component } from 'react'
import Layout from '../../Layout'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import PhotoGallery from '../../components/photoGallery/PhotoGallery'
import AlbumTitle from '../../components/AlbumTitle'
import { Checkbox } from 'semantic-ui-react'
import styled from 'styled-components'
import { authToken } from '../../authentication'
import PropTypes from 'prop-types'

const photoQuery = gql`
  query allPhotosPage($onlyWithFavorites: Boolean) {
    myAlbums(
      filter: { order_by: "title", order_direction: ASC, limit: 100 }
      onlyWithFavorites: $onlyWithFavorites
    ) {
      title
      id
      media(
        filter: { order_by: "media.title", order_direction: DESC, limit: 12 }
        onlyFavorites: $onlyWithFavorites
      ) {
        id
        title
        type
        thumbnail {
          url
          width
          height
        }
        highRes {
          url
          width
          height
        }
        videoWeb {
          url
        }
        favorite
      }
    }
  }
`

const FavoritesCheckbox = styled(Checkbox)`
  float: right;
  margin: 0.5rem 0 0 1rem;
`

class PhotosPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      activeAlbumIndex: -1,
      activePhotoIndex: -1,
      presenting: false,
      onlyWithFavorites: this.props.match.params.subPage === 'favorites',
    }

    this.setPresenting = this.setPresenting.bind(this)
    this.nextImage = this.nextImage.bind(this)
    this.previousImage = this.previousImage.bind(this)

    this.albums = []
  }

  onPopState(event) {
    this.state.setState({
      onlyWithFavorites: event.state.showFavorites,
    })
  }

  componentDidMount() {
    window.addEventListener('popstate', this.onPopState)
  }

  componentWillUnmount() {
    window.removeEventListener('popstate', this.onPopState)
  }

  favoritesCheckboxClick() {
    const onlyWithFavorites = !this.state.onlyWithFavorites
    history.pushState(
      { showFavorites: onlyWithFavorites },
      '',
      '/photos' + (onlyWithFavorites ? '/favorites' : '')
    )

    this.setState({
      onlyWithFavorites,
    })
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
    const albumImageCount = this.albums[this.state.activeAlbumIndex].media
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
    const showOnlyWithFavorites = this.state.onlyWithFavorites
    return (
      <Layout title="Photos">
        <Query
          query={photoQuery}
          variables={{ onlyWithFavorites: showOnlyWithFavorites }}
        >
          {({ loading, error, data }) => {
            if (error) return error

            if (loading) return null

            let galleryGroups = []
            let favoritesSwitch = ''

            this.albums = data.myAlbums

            if (data.myAlbums && authToken()) {
              favoritesSwitch = (
                <FavoritesCheckbox
                  toggle
                  label="Show only the favorites"
                  onClick={e => e.stopPropagation()}
                  checked={showOnlyWithFavorites}
                  onChange={() => {
                    this.favoritesCheckboxClick()
                  }}
                />
              )
              galleryGroups = data.myAlbums.map((album, index) => (
                <div key={album.id}>
                  <AlbumTitle album={album} />
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
                    media={album.media}
                    nextImage={this.nextImage}
                    previousImage={this.previousImage}
                  />
                </div>
              ))
            }

            return (
              <div>
                {favoritesSwitch}
                {galleryGroups}
              </div>
            )
          }}
        </Query>
      </Layout>
    )
  }
}

PhotosPage.propTypes = {
  match: PropTypes.shape({
    params: PropTypes.shape({
      subPage: PropTypes.string,
    }),
  }),
}

export default PhotosPage
