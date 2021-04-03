import PropTypes from 'prop-types'
import React, { useEffect } from 'react'
import { useLazyQuery, gql } from '@apollo/client'
import styled from 'styled-components'
import { authToken } from '../../helpers/authentication'
import { ProtectedImage, ProtectedVideo } from '../photoGallery/ProtectedMedia'
import SidebarShare from './Sharing'
import SidebarDownload from './SidebarDownload'
import SidebarItem from './SidebarItem'
import { SidebarFacesOverlay } from '../facesOverlay/FacesOverlay'

const mediaQuery = gql`
  query sidebarPhoto($id: ID!) {
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
      videoMetadata {
        id
        width
        height
        duration
        codec
        framerate
        bitrate
        colorProfile
        audio
      }
      exif {
        id
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
      faces {
        id
        rectangle {
          minX
          maxX
          minY
          maxY
        }
        faceGroup {
          id
        }
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

const PreviewVideo = styled(ProtectedVideo)`
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
`

const PreviewMedia = ({ media, previewImage }) => {
  if (media.type == null || media.type == 'photo') {
    return <PreviewImage src={previewImage?.url} />
  }

  if (media.type == 'video') {
    return <PreviewVideo media={media} />
  }

  throw new Error('Unknown media type')
}

PreviewMedia.propTypes = {
  media: PropTypes.object.isRequired,
  previewImage: PropTypes.object,
}

const Name = styled.div`
  text-align: center;
  font-size: 1.2rem;
  margin: 0.75rem 0 1rem;
`

const MetadataInfo = styled.div`
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

// From https://exiftool.org/TagNames/EXIF.html
const exposurePrograms = {
  0: 'Not defined',
  1: 'Manual',
  2: 'Normal program',
  3: 'Aperture priority',
  4: 'Shutter priority',
  5: 'Creative program',
  6: 'Action program',
  7: 'Portrait mode',
  8: 'Landscape mode ',
  9: 'Bulb',
}

// From https://exiftool.org/TagNames/EXIF.html#Flash
const flash = {
  0x0: 'No Flash',
  0x1: 'Fired',
  0x5: 'Fired, Return not detected',
  0x7: 'Fired, Return detected',
  0x8: 'On, Did not fire',
  0x9: 'On, Fired',
  0xd: 'On, Return not detected',
  0xf: 'On, Return detected',
  0x10: 'Off, Did not fire',
  0x14: 'Off, Did not fire, Return not detected',
  0x18: 'Auto, Did not fire',
  0x19: 'Auto, Fired',
  0x1d: 'Auto, Fired, Return not detected',
  0x1f: 'Auto, Fired, Return detected',
  0x20: 'No flash function',
  0x30: 'Off, No flash function',
  0x41: 'Fired, Red-eye reduction',
  0x45: 'Fired, Red-eye reduction, Return not detected',
  0x47: 'Fired, Red-eye reduction, Return detected',
  0x49: 'On, Red-eye reduction',
  0x4d: 'On, Red-eye reduction, Return not detected',
  0x4f: 'On, Red-eye reduction, Return detected',
  0x50: 'Off, Red-eye reduction',
  0x58: 'Auto, Did not fire, Red-eye reduction',
  0x59: 'Auto, Fired, Red-eye reduction',
  0x5d: 'Auto, Fired, Red-eye reduction, Return not detected',
  0x5f: 'Auto, Fired, Red-eye reduction, Return detected',
}

// From https://exiftool.org/TagNames/EXIF.html
// const orientation = {
//   1: 'Horizontal (normal)',
//   2: 'Mirror horizontal',
//   3: 'Rotate 180',
//   4: 'Mirror vertical',
//   5: 'Mirror horizontal and rotate 270 CW',
//   6: 'Rotate 90 CW',
//   7: 'Mirror horizontal and rotate 90 CW',
//   8: 'Rotate 270 CW',
// }

const SidebarContent = ({ media, hidePreview }) => {
  let exifItems = []

  if (media && media.exif) {
    let exifKeys = Object.keys(exifNameLookup).filter(
      x => media.exif[x] !== undefined && x != '__typename'
    )

    let exif = exifKeys.reduce(
      (prev, curr) => ({
        ...prev,
        [curr]: media.exif[curr],
      }),
      {}
    )

    exif.dateShot = new Date(exif.dateShot).toLocaleString()

    if (
      exif.exposureProgram !== undefined &&
      exif.exposureProgram !== 0 &&
      exposurePrograms[exif.exposureProgram]
    ) {
      exif.exposureProgram = exposurePrograms[exif.exposureProgram]
    }

    if (exif.aperture !== undefined) {
      exif.aperture = `f/${exif.aperture}`
    }

    if (exif.focalLength !== undefined) {
      exif.focalLength = `${exif.focalLength}mm`
    }

    if (exif.flash !== undefined && flash[exif.flash]) {
      exif.flash = flash[exif.flash]
    }

    exifItems = exifKeys.map(key => (
      <SidebarItem key={key} name={exifNameLookup[key]} value={exif[key]} />
    ))
  }

  let videoMetadataItems = []
  if (media && media.videoMetadata) {
    let metadata = Object.keys(media.videoMetadata)
      .filter(x => !['id', '__typename', 'width', 'height'].includes(x))
      .reduce(
        (prev, curr) => ({
          ...prev,
          [curr]: media.videoMetadata[curr],
        }),
        {}
      )

    metadata = {
      dimensions: `${media.videoMetadata.width}x${media.videoMetadata.height}`,
      ...metadata,
    }

    videoMetadataItems = Object.keys(metadata).map(key => (
      <SidebarItem key={key} name={key} value={metadata[key]} />
    ))
  }

  let previewImage = null
  if (media) {
    if (media.highRes) previewImage = media.highRes
    else if (media.thumbnail) previewImage = media.thumbnail
  }

  const imageAspect = previewImage
    ? previewImage.height / previewImage.width
    : 3 / 2

  return (
    <div>
      {!hidePreview && (
        <PreviewImageWrapper imageAspect={imageAspect}>
          <PreviewMedia previewImage={previewImage} media={media} />
          <SidebarFacesOverlay media={media} />
        </PreviewImageWrapper>
      )}
      <Name>{media && media.title}</Name>
      <MetadataInfo>{videoMetadataItems}</MetadataInfo>
      <MetadataInfo>{exifItems}</MetadataInfo>
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
    if (media != null && authToken()) {
      loadMedia({
        variables: {
          id: media.id,
        },
      })
    }
  }, [media])

  if (!media) return null

  if (!authToken()) {
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
