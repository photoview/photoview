import React from 'react'
import styled from 'styled-components'
import { MediaType } from '../../../../__generated__/globalTypes'
import { exhaustiveCheck } from '../../../helpers/utils'
import {
  ProtectedImage,
  ProtectedVideo,
  ProtectedVideoProps_Media,
} from '../ProtectedMedia'

const StyledPhoto = styled(ProtectedImage)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  object-fit: contain;
  object-position: center;
`

const StyledVideo = styled(ProtectedVideo)`
  position: absolute;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
`

export interface PresentMediaProps_Media extends ProtectedVideoProps_Media {
  type: MediaType
  highRes: null | {
    __typename: 'MediaURL'
    url: string
  }
}

type PresentMediaProps = {
  media: PresentMediaProps_Media
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
            src={media.thumbnail?.url}
            data-testid="present-img-thumbnail"
          />
          <StyledPhoto
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
