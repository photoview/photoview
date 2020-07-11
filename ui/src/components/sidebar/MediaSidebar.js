import React, { Component, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { useLazyQuery } from 'react-apollo'
import gql from 'graphql-tag'
import SidebarItem from './SidebarItem'
import ProtectedImage from '../photoGallery/ProtectedImage'
import SidebarShare from './Sharing'
import SidebarDownload from './SidebarDownload'

const mediaQuery = gql`
  query sidebarPhoto($id: Int!) {
    media(id: $id) {
      id
      title
      type
      highRes {
        url
        width
        height
      }
      thumbnail {
        url
        width
        height
      }
      videoWeb {
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

const PreviewVideo = styled.video`
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
`

const PreviewMedia = ({ media, previewImage }) => {
  if (media.type == null || media.type == 'photo') {
    return <PreviewImage src={previewImage.url} />
  }

  if (media.type == 'video') {
    return (
      <PreviewVideo controls key={media.id}>
        <source src={media.videoWeb.url} type="video/mp4" />
      </PreviewVideo>
    )
  }

  throw new Error('Unknown media type')
}

PreviewMedia.propTypes = {
  media: PropTypes.object.isRequired,
  previewImage: PropTypes.object.isRequired,
}

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

const SidebarContent = ({ media, hidePreview }) => {
  let exifItems = []

  if (media && media.exif) {
    let exifKeys = Object.keys(exifNameLookup).filter(
      x => !!media.exif[x] && x != '__typename'
    )

    let exif = exifKeys.reduce(
      (prev, curr) => ({
        ...prev,
        [curr]: media.exif[curr],
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
  if (media) {
    if (media.highRes) previewImage = media.highRes
    else if (media.thumbnail) previewImage = media.thumbnail
  }

  return (
    <div>
      {!hidePreview && (
        <PreviewImageWrapper
          imageAspect={previewImage.height / previewImage.width}
        >
          <PreviewMedia previewImage={previewImage} media={media} />
        </PreviewImageWrapper>
      )}
      <Name>{media && media.title}</Name>
      <ExifInfo>{exifItems}</ExifInfo>
      <SidebarDownload photo={media} />
      <SidebarShare photo={media} />
    </div>
  )
}

SidebarContent.propTypes = {
  media: PropTypes.object,
  hidePreview: PropTypes.bool,
}

const MediaSidebar = ({ media, hidePreview }) => {
  const [loadMedia, { loading, error, data }] = useLazyQuery(mediaQuery)

  useEffect(() => {
    if (media != null && localStorage.getItem('token')) {
      loadMedia({
        variables: {
          id: media.id,
        },
      })
    }
  }, [media])

  if (!media) return null

  if (!localStorage.getItem('token')) {
    return <SidebarContent media={media} hidePreview={hidePreview} />
  }

  if (error) return error

  if (loading || data == null) {
    return <SidebarContent media={media} hidePreview={hidePreview} />
  }

  return <SidebarContent media={data.media} hidePreview={hidePreview} />
}

MediaSidebar.propTypes = {
  media: PropTypes.object.isRequired,
  hidePreview: PropTypes.bool,
}

export default MediaSidebar
