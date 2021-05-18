import React from 'react'
import { Link } from 'react-router-dom'
import styled from 'styled-components'
import { MediaThumbnail } from '../photoGallery/MediaThumbnail'
import {
  toggleFavoriteAction,
  useMarkFavoriteMutation,
} from '../photoGallery/photoGalleryMutations'
import {
  getActiveTimelineImage,
  openTimelinePresentMode,
  TimelineGalleryAction,
  TimelineGalleryState,
} from './timelineGalleryReducer'

const MediaWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  height: 210px;
  position: relative;
  margin: -4px;
  padding-right: 16px;

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

const TotalItemsBubble = styled(Link)`
  position: absolute;
  top: 24px;
  right: 6px;
  background-color: white;
  border-radius: 50%;
  padding: 8px 0;
  box-shadow: 1px 1px 4px rgba(0, 0, 0, 0.3);
  color: black;
  width: 36px;
  height: 36px;
  text-align: center;
`

type TimelineGroupAlbumProps = {
  dateIndex: number
  albumIndex: number
  mediaState: TimelineGalleryState
  dispatchMedia: React.Dispatch<TimelineGalleryAction>
}

const TimelineGroupAlbum = ({
  dateIndex,
  albumIndex,
  mediaState,
  dispatchMedia,
}: TimelineGroupAlbumProps) => {
  const { media, mediaTotal, album } =
    mediaState.timelineGroups[dateIndex].groups[albumIndex]

  const [markFavorite] = useMarkFavoriteMutation()

  const mediaElms = media.map((media, index) => (
    <MediaThumbnail
      key={media.id}
      media={media}
      selectImage={() => {
        dispatchMedia({
          type: 'selectImage',
          index: {
            album: albumIndex,
            date: dateIndex,
            media: index,
          },
        })
      }}
      clickPresent={() => {
        openTimelinePresentMode({
          dispatchMedia,
          activeIndex: {
            album: albumIndex,
            date: dateIndex,
            media: index,
          },
        })
      }}
      clickFavorite={() => {
        toggleFavoriteAction({
          media,
          markFavorite,
        }).then(() => {
          dispatchMedia({
            type: 'selectImage',
            index: {
              album: albumIndex,
              date: dateIndex,
              media: index,
            },
          })
        })
      }}
      active={media.id === getActiveTimelineImage({ mediaState })?.id}
    />
  ))

  let itemsBubble = null
  const mediaVisibleCount = media.length
  if (mediaTotal > mediaVisibleCount) {
    itemsBubble = (
      <TotalItemsBubble to={`/album/${album.id}`}>
        {`+${Math.min(mediaTotal - mediaVisibleCount, 99)}`}
      </TotalItemsBubble>
    )
  }

  return (
    <GroupAlbumWrapper>
      <AlbumTitle>
        <Link to={`/album/${album.id}`}>{album.title}</Link>
      </AlbumTitle>
      <MediaWrapper>
        {mediaElms}
        {itemsBubble}
      </MediaWrapper>
    </GroupAlbumWrapper>
  )
}

export default TimelineGroupAlbum
