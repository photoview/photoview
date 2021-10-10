import React, { useContext, useEffect } from 'react'
import styled from 'styled-components'
import Layout from '../../components/layout/Layout'
import {
  ProtectedImage,
  ProtectedVideo,
} from '../../components/photoGallery/ProtectedMedia'
import { SidebarContext } from '../../components/sidebar/Sidebar'
import MediaSidebar from '../../components/sidebar/MediaSidebar/MediaSidebar'
import { useTranslation } from 'react-i18next'
import { SharePageToken_shareToken_media } from './__generated__/SharePageToken'
import { MediaType } from '../../__generated__/globalTypes'
import { exhaustiveCheck } from '../../helpers/utils'

const DisplayPhoto = styled(ProtectedImage)`
  /* width: 100%; */
  max-height: calc(80vh);
  object-fit: contain;
`

const DisplayVideo = styled(ProtectedVideo)`
  /* width: 100%; */
  max-height: calc(80vh);
`

type MediaViewProps = {
  media: SharePageToken_shareToken_media
}

const MediaView = ({ media }: MediaViewProps) => {
  const { updateSidebar } = useContext(SidebarContext)

  useEffect(() => {
    updateSidebar(<MediaSidebar media={media} hidePreview />)
  }, [media])

  switch (media.type) {
    case MediaType.Photo:
      return <DisplayPhoto src={media.highRes?.url} />
    case MediaType.Video:
      return <DisplayVideo media={media} />
  }

  exhaustiveCheck(media.type)
}

type MediaSharePageType = {
  media: SharePageToken_shareToken_media
}

const MediaSharePage = ({ media }: MediaSharePageType) => {
  const { t } = useTranslation()

  return (
    <Layout title={t('share_page.media.title', 'Shared media')}>
      <div data-testid="MediaSharePage">
        <h1 className="font-semibold text-xl mb-4">{media.title}</h1>
        <MediaView media={media} />
      </div>
    </Layout>
  )
}

export default MediaSharePage
