import React from 'react'
import styled from 'styled-components'
import { MediaType } from '../../../__generated__/globalTypes'
import { exhaustiveCheck } from '../../../helpers/utils'
import { ProtectedImage, ProtectedVideo } from '../ProtectedMedia'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'

// Sized to the presentation frame rather than the viewport: the frame is the
// full viewport until the info panel is pinned beside it, and then the photo
// has to fit the space that is left instead of running underneath the panel.
const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  object-position: center;
`

const StyledVideo = styled(ProtectedVideo)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
`

type PresentMediaProps = {
  media: MediaGalleryFields
  imageLoaded?(): void
}

const PresentMedia = ({
  media,
  imageLoaded,
  ...otherProps
}: PresentMediaProps) => {
  switch (media.type) {
    case MediaType.Photo:
      return (
        <div {...otherProps}>
          <StyledPhoto
            key={`${media.id}-thumb`}
            src={media.thumbnail?.url}
            data-testid="present-img-thumbnail"
          />
          <StyledPhoto
            key={`${media.id}-highres`}
            style={{ display: 'none' }}
            src={media.highRes?.url}
            data-testid="present-img-highres"
            onLoad={e => {
              const elem = e.target as HTMLImageElement
              elem.style.display = 'initial'
              imageLoaded && imageLoaded()
            }}
          />
        </div>
      )
    case MediaType.Video:
      return <StyledVideo media={media} data-testid="present-video" />
  }

  exhaustiveCheck(media.type)
}

export default PresentMedia
