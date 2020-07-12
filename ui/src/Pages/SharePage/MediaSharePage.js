import React, { useContext, useEffect } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import Layout from '../../Layout'
import ProtectedImage from '../../components/photoGallery/ProtectedImage'
import { SidebarContext } from '../../components/sidebar/Sidebar'
import MediaSidebar from '../../components/sidebar/MediaSidebar'

const DisplayPhoto = styled(ProtectedImage)`
  width: 100%;
  max-height: calc(80vh);
  object-fit: contain;
`

const DisplayVideo = styled.video`
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
    return (
      <DisplayVideo controls key={media.id}>
        <source src={media.videoWeb.url} type="video/mp4" />
      </DisplayVideo>
    )
  }

  throw new Error(`Unsupported media type: ${media.type}`)
}

MediaView.propTypes = {
  media: PropTypes.object.isRequired,
}

const MediaSharePage = ({ media }) => {
  return (
    <Layout>
      <h1>{media.title}</h1>
      <MediaView media={media} />
    </Layout>
  )
}

MediaSharePage.propTypes = {
  media: PropTypes.object.isRequired,
}

export default MediaSharePage
