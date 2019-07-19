import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import {
  Gallery,
  Photo,
  PhotoFiller,
  PhotoContainer,
  PhotoOverlay,
} from './styledElements'
import Layout from '../../Layout'
import { Loader } from 'semantic-ui-react'
import AlbumSidebar from './AlbumSidebar'

const albumQuery = gql`
  query albumQuery($id: ID) {
    album(id: $id) {
      title
      photos {
        id
        thumbnail {
          path
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
    }

    this.setActiveImage = this.setActiveImage.bind(this)

    this.photoAmount = 1
    this.previousActive = false

    this.keyDownEvent = e => {
      const activeImage = this.state.activeImage
      if (activeImage != -1) {
        if (e.key == 'ArrowRight') {
          this.setActiveImage((activeImage + 1) % this.photoAmount)
        }

        if (e.key == 'ArrowLeft') {
          if (activeImage <= 0) {
            this.setActiveImage(this.photoAmount - 1)
          } else {
            this.setActiveImage(activeImage - 1)
          }
        }
      }
    }
  }

  componentDidMount() {
    document.addEventListener('keydown', this.keyDownEvent)
  }

  componentWillUnmount() {
    document.removeEventListener('keydown', this.keyDownEvent)
  }

  setActiveImage(index) {
    this.previousActive = this.state.activeImage != -1
    this.setState({
      activeImage: index,
    })
  }

  dismissPopup() {
    this.setState({
      activeImage: -1,
    })
  }

  render() {
    const albumId = this.props.match.params.id

    return (
      <Layout>
        <Query query={albumQuery} variables={{ id: albumId }}>
          {({ loading, error, data }) => {
            if (error) return <div>Error</div>

            let photos = null
            if (data.album) {
              this.photoAmount = data.album.photos.length

              const { activeImage } = this.state

              photos = data.album.photos.map((photo, index) => {
                const active = activeImage == index

                return (
                  <PhotoContainer
                    key={photo.id}
                    onClick={() => {
                      this.setActiveImage(index)
                    }}
                  >
                    <Photo src={photo.thumbnail.path} />
                    <PhotoOverlay active={active} />
                  </PhotoContainer>
                )
              })
            }

            return (
              <div>
                <h1>{data.album ? data.album.title : ''}</h1>
                <Gallery>
                  <Loader active={loading}>Loading images</Loader>
                  {photos}
                  <PhotoFiller />
                </Gallery>
                <AlbumSidebar
                  imageId={
                    this.state.activeImage != -1
                      ? data.album.photos[this.state.activeImage].id
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
