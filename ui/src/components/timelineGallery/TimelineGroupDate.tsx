import React from 'react'
import TimelineGroupAlbum from './TimelineGroupAlbum'
import styled from 'styled-components'
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
const GroupDateWrapper = styled.div`
  margin: 12px 12px;
`

const DateTitle = styled.h1`
  font-size: 1.5rem;
  margin: 0 0 -12px;
`

const GroupAlbumWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  margin: 0 -8px;
`

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

  const albumGroupElms = group.groups.map((group, i) => (
    <TimelineGroupAlbum
      key={`${group.date}_${group.album.id}`}
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
      <div className="flex wrap -mx-2 my-0">{albumGroupElms}</div>
    </div>
  )
}

export default TimelineGroupDate
