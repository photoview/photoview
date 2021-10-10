import { gql, useLazyQuery } from '@apollo/client'
import React, { useEffect } from 'react'
import styled from 'styled-components'
import { authToken } from '../../../helpers/authentication'
import { MediaType } from '../../../__generated__/globalTypes'
import { SidebarFacesOverlay } from '../../facesOverlay/FacesOverlay'
import {
  ProtectedImage,
  ProtectedVideo,
  ProtectedVideoProps_Media,
} from '../../photoGallery/ProtectedMedia'
import { SidebarPhotoCover } from '../AlbumCovers'
import { SidebarPhotoShare } from '../Sharing'
import SidebarMediaDownload from '../SidebarDownloadMedia'
import SidebarHeader from '../SidebarHeader'
import { sidebarDownloadQuery_media_downloads } from '../__generated__/sidebarDownloadQuery'
import {
  sidebarPhoto,
  sidebarPhotoVariables,
  sidebarPhoto_media_exif,
  sidebarPhoto_media_faces,
  sidebarPhoto_media_thumbnail,
  sidebarPhoto_media_videoMetadata,
} from '../__generated__/sidebarPhoto'
import ExifDetails from './MediaSidebarExif'
import MediaSidebarMap from './MediaSidebarMap'

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
        coordinates {
          latitude
          longitude
        }
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

  let sidebarMap = null
  const mediaCoordinates = media.exif?.coordinates
  if (mediaCoordinates) {
    sidebarMap = <MediaSidebarMap coordinates={mediaCoordinates} />
  }

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
      <ExifDetails media={media} />
      {sidebarMap}
      <SidebarMediaDownload media={media} />
      <SidebarPhotoShare id={media.id} />
      <div className="mt-8">
        <SidebarPhotoCover cover_id={media.id} />
      </div>
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
