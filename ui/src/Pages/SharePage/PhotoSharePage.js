import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import Layout from '../../Layout'
import ProtectedImage from '../../components/photoGallery/ProtectedImage'
import { SidebarConsumer } from '../../components/sidebar/Sidebar'
import PhotoSidebar from '../../components/sidebar/PhotoSidebar'

const DisplayPhoto = styled(ProtectedImage)`
  width: 100%;
  max-height: calc(80vh);
  object-fit: contain;
`

const AlbumSharePage = ({ photo }) => {
  return (
    <Layout>
      <SidebarConsumer>
        {({ updateSidebar }) => (
          <>
            <h1>{photo.title}</h1>
            <DisplayPhoto
              src={photo.original.url}
              onLoad={() => {
                updateSidebar(<PhotoSidebar photo={photo} hidePreview />)
              }}
            />
          </>
        )}
      </SidebarConsumer>
    </Layout>
  )
}

AlbumSharePage.propTypes = {
  photo: PropTypes.object,
}

export default AlbumSharePage
