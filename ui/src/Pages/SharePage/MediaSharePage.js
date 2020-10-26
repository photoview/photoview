import React, { useContext, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import Layout from '../../Layout'
import {
  ProtectedImage,
  ProtectedVideo,
} from '../../components/photoGallery/ProtectedMedia'
import { SidebarContext } from '../../components/sidebar/Sidebar'
import MediaSidebar from '../../components/sidebar/MediaSidebar'

const DisplayPhoto = styled(ProtectedImage)`
  width: 100%;
  max-height: calc(80vh);
  object-fit: contain;
`

const DisplayVideo = styled(ProtectedVideo)`
  width: 100%;
  max-height: calc(80vh);
`

const MediaView = ({ media }) => {
  const { updateSidebar } = useContext(SidebarContext)

  useEffect(() => {
    updateSidebar(<MediaSidebar media={media} hidePreview />)
  }, [media])

  if (media.type == 'photo') {
    return <DisplayPhoto src={media.highRes.url} />
  }

  if (media.type == 'video') {
    return <DisplayVideo media={media} />
  }

  throw new Error(`Unsupported media type: ${media.type}`)
}

MediaView.propTypes = {
  media: PropTypes.object.isRequired,
}

const MediaSharePage = ({ media }) => (
  <Layout data-testid="MediaSharePage">
    <h1>{media.title}</h1>
    <MediaView media={media} />
  </Layout>
)

MediaSharePage.propTypes = {
  media: PropTypes.object.isRequired,
}

export default MediaSharePage
