import React from 'react'
import TimelineGroupAlbum from './TimelineGroupAlbum'
import { useTranslation } from 'react-i18next'
import {
  TimelineGalleryAction,
  TimelineGalleryState,
} from './timelineGalleryReducer'

const dateFormatterOptions: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
}

type TimelineGroupDateProps = {
  groupIndex: number
  mediaState: TimelineGalleryState
  dispatchMedia: React.Dispatch<TimelineGalleryAction>
}

const TimelineGroupDate = ({
  groupIndex,
  mediaState,
  dispatchMedia,
}: TimelineGroupDateProps) => {
  const { i18n } = useTranslation()

  const group = mediaState.timelineGroups[groupIndex]

  const albumGroupElms = group.albums.map((album, i) => (
    <TimelineGroupAlbum
      key={`${group.date}_${album.id}`}
      dateIndex={groupIndex}
      albumIndex={i}
      mediaState={mediaState}
      dispatchMedia={dispatchMedia}
    />
  ))

  const dateFormatter = new Intl.DateTimeFormat(
    i18n.language,
    dateFormatterOptions
  )

  const formattedDate = dateFormatter.format(new Date(group.date))

  return (
    <div className="mx-3 mb-2">
      <div className="text-xl m-0 -mb-2">{formattedDate}</div>
      <div className="flex flex-wrap -mx-2 my-0">{albumGroupElms}</div>
    </div>
  )
}

export default TimelineGroupDate
