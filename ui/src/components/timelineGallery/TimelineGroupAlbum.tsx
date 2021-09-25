import React from 'react'
import { Link } from 'react-router-dom'
import { MediaThumbnail } from '../photoGallery/MediaThumbnail'
import { PhotoFiller } from '../photoGallery/PhotoGallery'
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
  const {
    media,
    title: albumTitle,
    id: albumID,
  } = mediaState.timelineGroups[dateIndex].albums[albumIndex]

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

  return (
    <div className="mx-2">
      <Link to={`/album/${albumID}`} className="hover:underline">
        {albumTitle}
      </Link>
      <div className="flex flex-wrap items-center relative -mx-1 pr-4 overflow-hidden">
        {mediaElms}
        <PhotoFiller />
      </div>
    </div>
  )
}

export default TimelineGroupAlbum
