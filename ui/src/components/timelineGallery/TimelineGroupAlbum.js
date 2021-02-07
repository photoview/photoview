import React, { useContext } from 'react'
import PropTypes from 'prop-types'
import { MediaThumbnail } from '../photoGallery/MediaThumbnail'
import styled from 'styled-components'
import { SidebarContext } from '../sidebar/Sidebar'
import MediaSidebar from '../sidebar/MediaSidebar'
import { Link } from 'react-router-dom'

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
  font-size: 1.25rem;
  font-weight: 200;
  margin: 0 0 4px;

  & a:not(:hover) {
    color: #212121;
  }
`

const GroupAlbumWrapper = styled.div`
  margin: 12px 8px 0;
`

const TimelineGroupAlbum = ({
  group: { album, media /* mediaTotal */ },
  onSelectMedia,
  setPresenting,
  activeIndex,
  onFavorite,
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
      onFavorite={onFavorite}
      index={i}
      active={activeIndex == i}
    />
  ))

  return (
    <GroupAlbumWrapper>
      <AlbumTitle>
        <Link to={`/album/${album.id}`}>{album.title}</Link>
      </AlbumTitle>
      <MediaWrapper>{mediaElms}</MediaWrapper>
    </GroupAlbumWrapper>
  )
}

TimelineGroupAlbum.propTypes = {
  group: PropTypes.object.isRequired,
  onSelectMedia: PropTypes.func.isRequired,
  setPresenting: PropTypes.func.isRequired,
  activeIndex: PropTypes.number.isRequired,
  onFavorite: PropTypes.func,
}

export default TimelineGroupAlbum
