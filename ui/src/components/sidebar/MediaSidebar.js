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
import { isNil } from '../../helpers/utils'
import { useTranslation } from 'react-i18next'

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

const MetadataInfoContainer = styled.div`
  margin-bottom: 1.5rem;
`

export const MetadataInfo = ({ media }) => {
  const { t } = useTranslation()
  let exifItems = []

  const exifName = exifNameLookup(t)

  if (media?.exif) {
    let exifKeys = Object.keys(exifName).filter(
      x => media.exif[x] !== null && x != '__typename'
    )

    let exif = exifKeys.reduce(
      (prev, curr) => ({
        ...prev,
        [curr]: media.exif[curr],
      }),
      {}
    )

    if (!isNil(exif.dateShot)) {
      exif.dateShot = new Date(exif.dateShot).toLocaleString()
    }

    if (!isNil(exif.exposure) && exif.exposure !== 0) {
      exif.exposure = `1/${1 / exif.exposure}`
    }

    const exposurePrograms = exposureProgramsLookup(t)

    if (
      !isNil(exif.exposureProgram) &&
      exposurePrograms[exif.exposureProgram]
    ) {
      exif.exposureProgram = exposurePrograms[exif.exposureProgram]
    } else if (exif.exposureProgram !== 0) {
      delete exif.exposureProgram
    }

    if (!isNil(exif.aperture)) {
      exif.aperture = `f/${exif.aperture}`
    }

    if (!isNil(exif.focalLength)) {
      exif.focalLength = `${exif.focalLength}mm`
    }

    const flash = flashLookup(t)
    if (!isNil(exif.flash) && flash[exif.flash]) {
      exif.flash = flash[exif.flash]
    }

    exifItems = exifKeys.map(key => (
      <SidebarItem key={key} name={exifName[key]} value={exif[key]} />
    ))
  }

  let videoMetadataItems = []
  if (media?.videoMetadata) {
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

  return (
    <div>
      <MetadataInfoContainer>{videoMetadataItems}</MetadataInfoContainer>
      <MetadataInfoContainer>{exifItems}</MetadataInfoContainer>
    </div>
  )
}

MetadataInfo.propTypes = {
  media: PropTypes.object,
}

const exifNameLookup = t => ({
  camera: t('sidebar.media.exif.name.camera', 'Camera'),
  maker: t('sidebar.media.exif.name.maker', 'Maker'),
  lens: t('sidebar.media.exif.name.lens', 'Lens'),
  exposureProgram: t('sidebar.media.exif.name.exposure_program', 'Program'),
  dateShot: t('sidebar.media.exif.name.date_shot', 'Date shot'),
  exposure: t('sidebar.media.exif.name.exposure', 'Exposure'),
  aperture: t('sidebar.media.exif.name.aperture', 'Aperture'),
  iso: t('sidebar.media.exif.name.iso', 'ISO'),
  focalLength: t('sidebar.media.exif.name.focal_length', 'Focal length'),
  flash: t('sidebar.media.exif.name.flash', 'Flash'),
})

// From https://exiftool.org/TagNames/EXIF.html
const exposureProgramsLookup = t => ({
  0: t('sidebar.media.exif.exposure_program.not_defined', 'Not defined'),
  1: t('sidebar.media.exif.exposure_program.manual', 'Manual'),
  2: t('sidebar.media.exif.exposure_program.normal_program', 'Normal program'),
  3: t(
    'sidebar.media.exif.exposure_program.aperture_priority',
    'Aperture priority'
  ),
  4: t(
    'sidebar.media.exif.exposure_program.shutter_priority',
    'Shutter priority'
  ),
  5: t(
    'sidebar.media.exif.exposure_program.creative_program',
    'Creative program'
  ),
  6: t('sidebar.media.exif.exposure_program.action_program', 'Action program'),
  7: t('sidebar.media.exif.exposure_program.portrait_mode', 'Portrait mode'),
  8: t('sidebar.media.exif.exposure_program.landscape_mode', 'Landscape mode'),
  9: t('sidebar.media.exif.exposure_program.bulb', 'Bulb'),
})

// From https://exiftool.org/TagNames/EXIF.html#Flash
const flashLookup = t => {
  const values = {
    no_flash: t('sidebar.media.exif.flash.no_flash', 'No Flash'),
    fired: t('sidebar.media.exif.flash.fired', 'Fired'),
    did_not_fire: t('sidebar.media.exif.flash.did_not_fire', 'Did not fire'),
    on: t('sidebar.media.exif.flash.on', 'On'),
    off: t('sidebar.media.exif.flash.off', 'Off'),
    auto: t('sidebar.media.exif.flash.auto', 'Auto'),
    return_not_detected: t(
      'sidebar.media.exif.flash.return_not_detected',
      'Return not detected'
    ),
    return_detected: t(
      'sidebar.media.exif.flash.return_detected',
      'Return detected'
    ),
    no_flash_function: t(
      'sidebar.media.exif.flash.no_flash_function',
      'No flash function'
    ),
    red_eye_reduction: t(
      'sidebar.media.exif.flash.red_eye_reduction',
      'Red-eye reduction'
    ),
  }

  return {
    0x0: values['no_flash'],
    0x1: values['fired'],
    0x5: `${values['fired']}, ${values['return_not_detected']}`,
    0x7: `${values['fired']}, ${values['return_detected']}`,
    0x8: `${values['on']}, ${values['did_not_fire']}`,
    0x9: `${values['on']}, ${values['fired']}`,
    0xd: `${values['on']}, ${values['return_not_detected']}`,
    0xf: `${values['on']}, ${values['return_detected']}`,
    0x10: `${values['off']}, ${values['did_not_fire']}`,
    0x14: `${values['off']}, ${values['did_not_fire']}, ${values['return_not_detected']}`,
    0x18: `${values['auto']}, ${values['did_not_fire']}`,
    0x19: `${values['auto']}, ${values['fired']}`,
    0x1d: `${values['auto']}, ${values['fired']}, ${values['return_not_detected']}`,
    0x1f: `${values['auto']}, ${values['fired']}, ${values['return_detected']}`,
    0x20: `${values['no_flash_function']}`,
    0x30: `${values['off']}, ${values['no_flash_function']}`,
    0x41: `${values['fired']}, ${values['red_eye_reduction']}`,
    0x45: `${values['fired']}, ${values['red_eye_reduction']}, ${values['return_not_detected']}`,
    0x47: `${values['fired']}, ${values['red_eye_reduction']}, ${values['return_detected']}`,
    0x49: `${values['on']}, ${values['red_eye_reduction']}`,
    0x4d: `${values['on']}, ${values['red_eye_reduction']}, ${values['return_not_detected']}`,
    0x4f: `${values['on']}, ${values['red_eye_reduction']}, ${values['return_detected']}`,
    0x50: `${values['off']}, ${values['red_eye_reduction']}`,
    0x58: `${values['auto']}, ${values['did_not_fire']}, ${values['red_eye_reduction']}`,
    0x59: `${values['auto']}, ${values['fired']}, ${values['red_eye_reduction']}`,
    0x5d: `${values['auto']}, ${values['red_eye_reduction']}, ${values['return_not_detected']}`,
    0x5f: `${values['auto']}, ${values['red_eye_redcution']}, ${values['return_detected']}`,
  }
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
      <MetadataInfo media={media} />
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
