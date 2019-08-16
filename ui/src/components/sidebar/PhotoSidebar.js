import React, { Component } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import SidebarItem from './SidebarItem'
import { Loader } from 'semantic-ui-react'
import ProtectedImage from '../photoGallery/ProtectedImage'
import SidebarShare from './Sharing'

const photoQuery = gql`
  query sidebarPhoto($id: ID!) {
    photo(id: $id) {
      id
      title
      original {
        url
        width
        height
      }
      exif {
        camera
        maker
        lens
        dateShot {
          formatted
        }
        fileSize
        exposure
        aperture
        iso
        focalLength
        flash
      }
    }
  }
`

const PreviewImage = styled(ProtectedImage)`
  width: 100%;
  height: 333px;
  object-fit: contain;
`

const Name = styled.div`
  text-align: center;
  font-size: 16px;
  margin-bottom: 12px;
`

const exifNameLookup = {
  camera: 'Camera',
  maker: 'Maker',
  lens: 'Lens',
  dateShot: 'Date Shot',
  fileSize: 'File Size',
  exposure: 'Exposure',
  aperture: 'Aperture',
  iso: 'ISO',
  focalLength: 'Focal Length',
  flash: 'Flash',
}

class PhotoSidebar extends Component {
  render() {
    const { imageId } = this.props

    if (!imageId) {
      return null
    }

    return (
      <div>
        <Query query={photoQuery} variables={{ id: imageId }}>
          {({ loading, error, data }) => {
            if (error) return error

            const { photo } = data
            let exifItems = []

            if (photo && photo.exif) {
              let exifKeys = Object.keys(photo.exif).filter(
                x => !!photo.exif[x] && x != '__typename'
              )

              let exif = exifKeys.reduce(
                (prev, curr) => ({
                  ...prev,
                  [curr]: photo.exif[curr],
                }),
                {}
              )

              exif.dateShot = new Date(exif.dateShot.formatted).toLocaleString()

              exifItems = exifKeys.map(key => (
                <SidebarItem
                  key={key}
                  name={exifNameLookup[key]}
                  value={exif[key]}
                />
              ))
            }

            return (
              <div>
                <Loader active={loading} />
                <PreviewImage
                  src={photo && photo.original && photo.original.url}
                />
                <Name>{photo && photo.title}</Name>
                <div>{exifItems}</div>
                <SidebarShare photo={photo} />
              </div>
            )
          }}
        </Query>
      </div>
    )
  }
}

PhotoSidebar.propTypes = {
  imageId: PropTypes.string.isRequired,
}

export default PhotoSidebar
