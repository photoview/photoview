import React, { Component } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Query } from 'react-apollo'
import gql from 'graphql-tag'
import SidebarItem from './SidebarItem'
import ProtectedImage from '../photoGallery/ProtectedImage'
import SidebarShare from './Sharing'
import SidebarDownload from './SidebarDownload'

const photoQuery = gql`
  query sidebarPhoto($id: Int!) {
    media(id: $id) {
      id
      title
      highRes {
        url
        width
        height
      }
      exif {
        camera
        maker
        lens
        dateShot
        exposure
        aperture
        iso
        focalLength
        flash
        exposureProgram
      }
    }
  }
`

const PreviewImageWrapper = styled.div`
  width: 100%;
  height: 0;
  padding-top: ${({ imageAspect }) => Math.min(imageAspect, 0.75) * 100}%;
  position: relative;
`

const PreviewImage = styled(ProtectedImage)`
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  object-fit: contain;
`

const Name = styled.div`
  text-align: center;
  font-size: 1.2rem;
  margin: 0.75rem 0 1rem;
`

const ExifInfo = styled.div`
  margin-bottom: 1.5rem;
`

const exifNameLookup = {
  camera: 'Camera',
  maker: 'Maker',
  lens: 'Lens',
  exposureProgram: 'Program',
  dateShot: 'Date Shot',
  exposure: 'Exposure',
  aperture: 'Aperture',
  iso: 'ISO',
  focalLength: 'Focal Length',
  flash: 'Flash',
}

const exposurePrograms = {
  '0': 'Not defined',
  '1': 'Manual',
  '2': 'Normal program',
  '3': 'Aperture priority',
  '4': 'Shutter priority',
  '5': 'Creative program',
  '6': 'Action program',
  '7': 'Portrait mode',
  '8': 'Landscape mode ',
}

const SidebarContent = ({ photo, hidePreview }) => {
  let exifItems = []

  if (photo && photo.exif) {
    let exifKeys = Object.keys(exifNameLookup).filter(
      x => !!photo.exif[x] && x != '__typename'
    )

    let exif = exifKeys.reduce(
      (prev, curr) => ({
        ...prev,
        [curr]: photo.exif[curr],
      }),
      {}
    )

    exif.dateShot = new Date(exif.dateShot).toLocaleString()
    if (exif.exposureProgram) {
      exif.exposureProgram = exposurePrograms[exif.exposureProgram]
    }

    if (exif.aperture) {
      exif.aperture = `f/${exif.aperture}`
    }

    if (exif.focalLength) {
      exif.focalLength = `${exif.focalLength}mm`
    }

    exif.exposureProgram = exifItems = exifKeys.map(key => (
      <SidebarItem key={key} name={exifNameLookup[key]} value={exif[key]} />
    ))
  }

  let previewImage = null
  if (photo) {
    if (photo.highRes) previewImage = photo.highRes
    else if (photo.thumbnail) previewImage = photo.thumbnail
  }

  return (
    <div>
      {!hidePreview && (
        <PreviewImageWrapper
          imageAspect={previewImage.height / previewImage.width}
        >
          <PreviewImage src={previewImage.url} />
        </PreviewImageWrapper>
      )}
      <Name>{photo && photo.title}</Name>
      <ExifInfo>{exifItems}</ExifInfo>
      <SidebarDownload photo={photo} />
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
              <SidebarContent photo={data.media} hidePreview={hidePreview} />
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
