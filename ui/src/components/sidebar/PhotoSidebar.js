import React, { Component } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import SidebarItem from './SidebarItem'
import { Loader } from 'semantic-ui-react'
import ProtectedImage from '../photoGallery/ProtectedImage'
import SidebarShare from './Sharing'
import { SidebarConsumer } from './Sidebar'

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

const SidebarContent = ({ photo, hidePreview }) => {
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
      <SidebarItem key={key} name={exifNameLookup[key]} value={exif[key]} />
    ))
  }

  let previewUrl = null
  if (photo) {
    if (photo.original) previewUrl = photo.original.url
    else if (photo.thumbnail) previewUrl = photo.thumbnail.url
  }

  return (
    <div>
      {!hidePreview && <PreviewImage src={previewUrl} />}
      <Name>{photo && photo.title}</Name>
      <div>{exifItems}</div>
      <SidebarShare photo={photo} />
    </div>
  )
}

SidebarContent.propTypes = {
  photo: PropTypes.object,
  hidePreview: PropTypes.bool,
}

class PhotoSidebar extends Component {
  render() {
    const { photo, hidePreview } = this.props

    if (!photo) return null

    if (!localStorage.getItem('token')) {
      return <SidebarContent photo={photo} hidePreview={hidePreview} />
    }

    return (
      <div>
        <Query query={photoQuery} variables={{ id: photo.id }}>
          {({ loading, error, data }) => {
            if (error) return error

            if (loading) {
              return <SidebarContent photo={photo} hidePreview={hidePreview} />
            }

            return (
              <SidebarContent photo={data.photo} hidePreview={hidePreview} />
            )
          }}
        </Query>
      </div>
    )
  }
}

PhotoSidebar.propTypes = {
  photo: PropTypes.object.isRequired,
  hidePreview: PropTypes.bool,
}

export default PhotoSidebar
