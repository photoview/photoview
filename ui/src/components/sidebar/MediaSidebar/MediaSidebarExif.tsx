import React from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { isNil } from '../../../helpers/utils'
import { TranslationFn } from '../../../localization'
import SidebarItem from '../SidebarItem'
import { MediaSidebarMedia } from './MediaSidebar'

const MetadataInfoContainer = styled.div`
  margin-bottom: 1.5rem;
`

type ExifDetailsProps = {
  media?: MediaSidebarMedia
}

const ExifDetails = ({ media }: ExifDetailsProps) => {
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

    const coordinates = media.exif.coordinates
    if (!isNil(coordinates)) {
      exif.coordinates = `${
        Math.round(coordinates.latitude * 1000000) / 1000000
      }, ${Math.round(coordinates.longitude * 1000000) / 1000000}`
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
        const value = videoMetadata[curr]
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
  description: t('sidebar.media.exif.description', 'Description'),
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
  coordinates: t('sidebar.media.exif.name.coordinates', 'Coordinates'),
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

export default ExifDetails
