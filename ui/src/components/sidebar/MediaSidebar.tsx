import React, { useEffect } from 'react'
import { useLazyQuery, gql } from '@apollo/client'
import styled from 'styled-components'
import { authToken } from '../../helpers/authentication'
import {
  ProtectedImage,
  ProtectedVideo,
  ProtectedVideoProps_Media,
} from '../photoGallery/ProtectedMedia'
import { SidebarPhotoShare } from './Sharing'
import SidebarDownload from './SidebarDownload'
import SidebarItem from './SidebarItem'
import { SidebarFacesOverlay } from '../facesOverlay/FacesOverlay'
import { isNil } from '../../helpers/utils'
import { useTranslation } from 'react-i18next'
import { MediaType } from '../../__generated__/globalTypes'
import { TranslationFn } from '../../localization'
import {
  sidebarPhoto,
  sidebarPhotoVariables,
  sidebarPhoto_media_exif,
  sidebarPhoto_media_faces,
  sidebarPhoto_media_thumbnail,
  sidebarPhoto_media_videoMetadata,
} from './__generated__/sidebarPhoto'

import { sidebarDownloadQuery_media_downloads } from './__generated__/sidebarDownloadQuery'
import SidebarHeader from './SidebarHeader'

const SIDEBAR_MEDIA_QUERY = gql`
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

interface PreviewMediaPropsMedia extends ProtectedVideoProps_Media {
  type: MediaType
}

type PreviewMediaProps = {
  media: PreviewMediaPropsMedia
  previewImage?: {
    url: string
  }
}

const PreviewMedia = ({ media, previewImage }: PreviewMediaProps) => {
  if (media.type === MediaType.Photo) {
    return <PreviewImage src={previewImage?.url} />
  }

  if (media.type === MediaType.Video) {
    return <PreviewVideo media={media} />
  }

  return <div>ERROR: Unknown media type: {media.type}</div>
}

const MetadataInfoContainer = styled.div`
  margin-bottom: 1.5rem;
`

type MediaInfoProps = {
  media?: MediaSidebarMedia
}

export const MetadataInfo = ({ media }: MediaInfoProps) => {
  const { t } = useTranslation()
  let exifItems: JSX.Element[] = []

  const exifName = exifNameLookup(t)

  if (media?.exif) {
    const mediaExif = media?.exif as unknown as {
      [key: string]: string | number | null
    }

    const exifKeys = Object.keys(exifName).filter(
      x => mediaExif[x] !== null && x != '__typename'
    )

    const exif = exifKeys.reduce((prev, curr) => {
      const value = mediaExif[curr]
      if (isNil(value)) return prev

      return {
        ...prev,
        [curr]: value,
      }
    }, {} as { [key: string]: string | number })

    if (!isNil(exif.dateShot)) {
      exif.dateShot = new Date(exif.dateShot).toLocaleString()
    }

    if (typeof exif.exposure === 'number' && exif.exposure !== 0) {
      exif.exposure = `1/${Math.round(1 / exif.exposure)}`
    }

    const exposurePrograms = exposureProgramsLookup(t)

    if (
      typeof exif.exposureProgram === 'number' &&
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
    if (typeof exif.flash === 'number' && flash[exif.flash]) {
      exif.flash = flash[exif.flash]
    }

    exifItems = exifKeys.map(key => (
      <SidebarItem key={key} name={exifName[key]} value={exif[key] as string} />
    ))
  }

  let videoMetadataItems: JSX.Element[] = []
  if (media?.videoMetadata) {
    const videoMetadata = media.videoMetadata as unknown as {
      [key: string]: string | number | null
    }

    let metadata = Object.keys(videoMetadata)
      .filter(x => !['id', '__typename', 'width', 'height'].includes(x))
      .reduce((prev, curr) => {
        const value = videoMetadata[curr as string]
        if (isNil(value)) return prev

        return {
          ...prev,
          [curr]: value,
        }
      }, {} as { [key: string]: string | number })

    metadata = {
      dimensions: `${media.videoMetadata.width}x${media.videoMetadata.height}`,
      ...metadata,
    }

    videoMetadataItems = Object.keys(metadata).map(key => (
      <SidebarItem key={key} name={key} value={metadata[key] as string} />
    ))
  }

  return (
    <div>
      <MetadataInfoContainer>{videoMetadataItems}</MetadataInfoContainer>
      <MetadataInfoContainer>{exifItems}</MetadataInfoContainer>
    </div>
  )
}

const exifNameLookup = (t: TranslationFn): { [key: string]: string } => ({
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
const exposureProgramsLookup = (
  t: TranslationFn
): { [key: number]: string } => ({
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
const flashLookup = (t: TranslationFn): { [key: number]: string } => {
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
    0x5f: `${values['auto']}, ${values['red_eye_reduction']}, ${values['return_detected']}`,
  }
}

type SidebarContentProps = {
  media: MediaSidebarMedia
  hidePreview?: boolean
}

const SidebarContent = ({ media, hidePreview }: SidebarContentProps) => {
  let previewImage = null
  if (media.highRes) previewImage = media.highRes
  else if (media.thumbnail) previewImage = media.thumbnail

  const imageAspect =
    previewImage?.width && previewImage?.height
      ? previewImage.height / previewImage.width
      : 3 / 2

  return (
    <div>
      <SidebarHeader title={media.title ?? 'Loading...'} />
      <div className="lg:mx-4">
        {!hidePreview && (
          <div
            className="w-full h-0 relative"
            style={{ paddingTop: `${Math.min(imageAspect, 0.75) * 100}%` }}
          >
            <PreviewMedia
              previewImage={previewImage || undefined}
              media={media}
            />
            <SidebarFacesOverlay media={media} />
          </div>
        )}
      </div>
      <MetadataInfo media={media} />
      <SidebarDownload media={media} />
      <SidebarPhotoShare id={media.id} />
    </div>
  )
}

export interface MediaSidebarMedia {
  __typename: 'Media'
  id: string
  title?: string
  type: MediaType
  highRes?: null | {
    __typename: 'MediaURL'
    url: string
    width?: number
    height?: number
  }
  thumbnail?: sidebarPhoto_media_thumbnail | null
  videoWeb?: null | {
    __typename: 'MediaURL'
    url: string
    width?: number
    height?: number
  }
  videoMetadata?: sidebarPhoto_media_videoMetadata | null
  exif?: sidebarPhoto_media_exif | null
  faces?: sidebarPhoto_media_faces[]
  downloads?: sidebarDownloadQuery_media_downloads[]
}

type MediaSidebarType = {
  media: MediaSidebarMedia
  hidePreview?: boolean
}

const MediaSidebar = ({ media, hidePreview }: MediaSidebarType) => {
  const [loadMedia, { loading, error, data }] = useLazyQuery<
    sidebarPhoto,
    sidebarPhotoVariables
  >(SIDEBAR_MEDIA_QUERY)

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

  if (error) return <div>{error.message}</div>

  if (loading || data == null) {
    return <SidebarContent media={media} hidePreview={hidePreview} />
  }

  return <SidebarContent media={data.media} hidePreview={hidePreview} />
}

export default MediaSidebar
