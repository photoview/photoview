import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import { Gallery, Photo, PhotoFiller } from './styledElements'
import Layout from '../../Layout'

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

    this.keyUpEvent = e => {
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
    document.addEventListener('keyup', this.keyUpEvent)
  }

  componentWillUnmount() {
    document.removeEventListener('keyup', this.keyUpEvent)
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
            if (loading) return <div>Loading</div>

            this.photoAmount = data.album.photos.length

            const { activeImage } = this.state

            const photos = data.album.photos.map((photo, index) => {
              return <Photo key={photo.id} src={photo.thumbnail.path}></Photo>
            })

            return (
              <div>
                <h1>{data.album.title}</h1>
                <Gallery>
                  {photos}
                  <PhotoFiller />
                </Gallery>
              </div>
            )
          }}
        </Query>
      </Layout>
    )
  }
}

export default AlbumPage
