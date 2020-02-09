import React, { Component } from 'react'
import ReactRouterPropTypes from 'react-router-prop-types'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import Layout from '../../Layout'
import PhotoGallery, {
  presentIndexFromHash,
} from '../../components/photoGallery/PhotoGallery'
import AlbumTitle from '../../components/AlbumTitle'
import AlbumGallery from '../../components/albumGallery/AlbumGallery'

const albumQuery = gql`
  query albumQuery($id: Int!) {
    album(id: $id) {
      id
      title
      subAlbums(filter: { order_by: "title" }) {
        id
        title
        thumbnail {
          thumbnail {
            url
          }
        }
      }
      photos(filter: { order_by: "title", order_direction: DESC }) {
        id
        thumbnail {
          url
          width
          height
        }
        original {
          url
        }
      }
    }
  }
`

class AlbumPage extends Component {
  // constructor(props) {
  //   super(props)

  //   this.state = {
  //     activeImage: -1,
  //     presenting: false,
  //   }

  //   const presentIndex = presentIndexFromHash(document.location.hash)
  //   if (presentIndex) {
  //     this.state.activeImage = presentIndex
  //     this.state.presenting = true
  //   }

  //   this.setActiveImage = this.setActiveImage.bind(this)
  //   this.nextImage = this.nextImage.bind(this)
  //   this.previousImage = this.previousImage.bind(this)
  //   this.setPresenting = this.setPresenting.bind(this)

  //   this.photos = []
  // }

  // setActiveImage(index) {
  //   this.setState({
  //     activeImage: index,
  //   })
  // }

  // nextImage() {
  //   this.setState({
  //     activeImage: (this.state.activeImage + 1) % this.photos.length,
  //   })
  // }

  // previousImage() {
  //   if (this.state.activeImage <= 0) {
  //     this.setState({
  //       activeImage: this.photos.length - 1,
  //     })
  //   } else {
  //     this.setState({
  //       activeImage: this.state.activeImage - 1,
  //     })
  //   }
  // }

  // componentDidUpdate(prevProps, prevState) {
  //   if (this.state.presenting) {
  //     window.history.replaceState(
  //       null,
  //       null,
  //       document.location.pathname + '#' + `present=${this.state.activeImage}`
  //     )
  //   } else if (presentIndexFromHash(document.location.hash)) {
  //     window.history.replaceState(
  //       null,
  //       null,
  //       document.location.pathname.split('#')[0]
  //     )
  //   }
  // }

  // setPresenting(presenting) {
  //   console.log('Presenting', presenting, this)
  //   this.setState({
  //     presenting,
  //   })
  // }

  render() {
    const albumId = this.props.match.params.id

    return (
      <Query query={albumQuery} variables={{ id: albumId }}>
        {({ loading, error, data }) => {
          if (error) return <div>Error</div>

          return <AlbumGallery album={data && data.album} loading={loading} />
        }}
      </Query>
    )
  }
}

AlbumPage.propTypes = {
  ...ReactRouterPropTypes,
}

export default AlbumPage
