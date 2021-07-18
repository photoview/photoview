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
    <div className="mx-2">
      <Link to={`/album/${album.id}`} className="hover:underline">
        {album.title}
      </Link>
      <div className="flex flex-wrap items-center h-[210px] relative -mx-1 pr-4 overflow-hidden">
        {mediaElms}
        {itemsBubble}
      </div>
    </div>
  )
}

export default TimelineGroupAlbum
