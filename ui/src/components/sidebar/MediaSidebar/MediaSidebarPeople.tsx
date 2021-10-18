import React from 'react'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { MediaSidebarMedia } from './MediaSidebar'
import { sidebarMediaQuery_media_faces } from './__generated__/sidebarMediaQuery'

type MediaSidebarFaceProps = {
  face: sidebarMediaQuery_media_faces
}

const MediaSidebarPerson = ({ face }: MediaSidebarFaceProps) => {
  return <div>{face.faceGroup.label ?? 'unlabeled'}</div>
}

type MediaSidebarFacesProps = {
  media: MediaSidebarMedia
}

const MediaSidebarPeople = ({ media }: MediaSidebarFacesProps) => {
  const { t } = useTranslation()
  const faceElms = (media.faces ?? []).map(face => (
    <MediaSidebarPerson key={face.id} face={face} />
  ))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.people.title', 'People')}
      </SidebarSectionTitle>
      <div>{faceElms}</div>
    </SidebarSection>
  )
}

export default MediaSidebarPeople
