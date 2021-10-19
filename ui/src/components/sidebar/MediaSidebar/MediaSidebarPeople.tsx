import React from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import FaceCircleImage from '../../../Pages/PeoplePage/FaceCircleImage'
import { Button } from '../../../primitives/form/Input'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { MediaSidebarMedia } from './MediaSidebar'
import { sidebarMediaQuery_media_faces } from './__generated__/sidebarMediaQuery'

import { ReactComponent as PeopleDotsIcon } from './icons/peopleDotsIcon.svg'

type MediaSidebarFaceProps = {
  face: sidebarMediaQuery_media_faces
}

const MediaSidebarPerson = ({ face }: MediaSidebarFaceProps) => {
  const { t } = useTranslation()

  return (
    <li className="inline-block">
      <Link to={`/people/${face.faceGroup.id}`}>
        <FaceCircleImage imageFace={face} selectable={true} size="100px" />
      </Link>
      <div
        className={`text-center ${
          face.faceGroup.label ? 'text-black' : 'text-gray-600'
        }`}
      >
        {face.faceGroup.label ??
          t('people_page.face_group.unlabeled', 'Unlabeled')}
        <Button className="px-2 py-1.5 align-middle ml-1">
          <PeopleDotsIcon className="text-gray-500" />
        </Button>
      </div>
    </li>
  )
}

type MediaSidebarFacesProps = {
  media: MediaSidebarMedia
}

const MediaSidebarPeople = ({ media }: MediaSidebarFacesProps) => {
  const { t } = useTranslation()

  const faceElms = (media.faces ?? []).map(face => (
    <MediaSidebarPerson key={face.id} face={face} />
  ))

  if (faceElms.length == 0) return null

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.people.title', 'People')}
      </SidebarSectionTitle>
      <div className="overflow-x-auto">
        <ul className="flex gap-4 mx-4">{faceElms}</ul>
      </div>
    </SidebarSection>
  )
}

export default MediaSidebarPeople
