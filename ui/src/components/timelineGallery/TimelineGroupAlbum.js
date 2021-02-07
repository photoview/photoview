import React, { useContext } from 'react'
import PropTypes from 'prop-types'
import { MediaThumbnail } from '../photoGallery/MediaThumbnail'
import styled from 'styled-components'
import { SidebarContext } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'

const MediaWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  height: 210px;
  position: relative;
  margin: -4px;

  overflow: hidden;

  @media (max-width: 1000px) {
    /* Compensate for tab bar on mobile */
    margin-bottom: 76px;
  }
`

const AlbumTitle = styled.h2`
  color: #212121;
  font-size: 1.25rem;
  font-weight: 200;
  margin: 0 0 4px;
`

const GroupAlbumWrapper = styled.div`
  margin-top: 12px;
`

const TimelineGroupAlbum = ({
  group: { album, media /* mediaTotal */ },
  onSelectMedia,
  setPresenting,
  activeIndex,
}) => {
  const { updateSidebar } = useContext(SidebarContext)

  const mediaElms = media.map((media, i) => (
    <MediaThumbnail
      key={media.id}
      media={media}
      onSelectImage={index => {
        onSelectMedia(index)
        updateSidebar(<MediaSidebar media={media} />)
      }}
      setPresenting={setPresenting}
      index={i}
      active={activeIndex == i}
    />
  ))

  return (
    <GroupAlbumWrapper>
      <AlbumTitle>{album.title}</AlbumTitle>
      <MediaWrapper>{mediaElms}</MediaWrapper>
    </GroupAlbumWrapper>
  )
}

TimelineGroupAlbum.propTypes = {
  group: PropTypes.object.isRequired,
  onSelectMedia: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
  activeIndex: PropTypes.number.isRequired,
}

export default TimelineGroupAlbum
