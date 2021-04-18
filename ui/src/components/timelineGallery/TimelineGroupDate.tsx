import React from 'react'
import TimelineGroupAlbum from './TimelineGroupAlbum'
import styled from 'styled-components'
import { myTimeline_myTimeline } from './__generated__/myTimeline'
import { TimelineActiveIndex } from './TimelineGallery'
import { useTranslation } from 'react-i18next'

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
  date: string
  groups: myTimeline_myTimeline[]
  onSelectDateGroup(args: { media: number; albumGroup: number }): void
  activeIndex: TimelineActiveIndex
  setPresenting: React.Dispatch<React.SetStateAction<boolean>>
  onFavorite(): void
}

const TimelineGroupDate = ({
  date,
  groups,
  onSelectDateGroup,
  activeIndex,
  setPresenting,
  onFavorite,
}: TimelineGroupDateProps) => {
  const { i18n } = useTranslation()

  const albumGroupElms = groups.map((group, i) => (
    <TimelineGroupAlbum
      key={`${group.date}_${group.album.id}`}
      group={group}
      onSelectMedia={mediaIndex => {
        onSelectDateGroup({
          media: mediaIndex,
          albumGroup: i,
        })
      }}
      activeIndex={activeIndex.albumGroup == i ? activeIndex.media : -1}
      setPresenting={setPresenting}
      onFavorite={onFavorite}
    />
  ))

  const dateFormatter = new Intl.DateTimeFormat(
    i18n.language,
    dateFormatterOptions
  )

  const formattedDate = dateFormatter.format(new Date(date))

  return (
    <GroupDateWrapper>
      <DateTitle>{formattedDate}</DateTitle>
      <GroupAlbumWrapper>{albumGroupElms}</GroupAlbumWrapper>
    </GroupDateWrapper>
  )
}

export default TimelineGroupDate
