import React from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import FaceCircleImage from '../../../Pages/PeoplePage/FaceCircleImage'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { MediaSidebarMedia } from './MediaSidebar'
import { sidebarMediaQuery_media_faces } from './__generated__/sidebarMediaQuery'

import { ReactComponent as PeopleDotsIcon } from './icons/peopleDotsIcon.svg'
import { Menu } from '@headlessui/react'
import { Button } from '../../../primitives/form/Input'
import { ArrowPopoverPanel } from '../Sharing'
import { tailwindClassNames } from '../../../helpers/utils'

type PersonMoreMenuItemProps = {
  label: string
  className?: string
}

const PersonMoreMenuItem = ({ label, className }: PersonMoreMenuItemProps) => {
  return (
    <Menu.Item>
      {({ active }) => (
        <a
          className={tailwindClassNames(
            `block py-1 cursor-pointer ${active && 'bg-gray-50 text-black'}`,
            className
          )}
        >
          {label}
        </a>
      )}
    </Menu.Item>
  )
}

type PersonMoreMenuProps = {
  face: sidebarMediaQuery_media_faces
}

const PersonMoreMenu = ({ face }: PersonMoreMenuProps) => {
  const { t } = useTranslation()

  face
  return (
    <Menu as="div" className="relative inline-block">
      <Menu.Button as={Button} className="px-1.5 py-1.5 align-middle ml-1">
        <PeopleDotsIcon className="text-gray-500" />
      </Menu.Button>
      <Menu.Items className="">
        <ArrowPopoverPanel width={120}>
          <PersonMoreMenuItem
            className="border-b"
            label={t('people_page.action_label.change_label', 'Change label')}
          />
          <PersonMoreMenuItem
            className="border-b"
            label={t('sidebar.people.action_label.merge_face', 'Merge face')}
          />
          <PersonMoreMenuItem
            className="border-b"
            label={t(
              'sidebar.people.action_label.detach_image',
              'Detach image'
            )}
          />
          <PersonMoreMenuItem
            label={t('sidebar.people.action_label.move_face', 'Move face')}
          />
        </ArrowPopoverPanel>
      </Menu.Items>
    </Menu>
  )
}

type MediaSidebarFaceProps = {
  face: sidebarMediaQuery_media_faces
}

const MediaSidebarPerson = ({ face }: MediaSidebarFaceProps) => {
  const { t } = useTranslation()

  return (
    <li className="inline-block">
      <Link to={`/people/${face.faceGroup.id}`}>
        <FaceCircleImage imageFace={face} selectable={true} size="92px" />
      </Link>
      <div
        className={`text-center text-sm mt-1 ${
          face.faceGroup.label ? 'text-black' : 'text-gray-600'
        }`}
      >
        {face.faceGroup.label ??
          t('people_page.face_group.unlabeled', 'Unlabeled')}
        <PersonMoreMenu face={face} />
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
      <div
        className="overflow-x-auto mb-[-200px]"
        style={{ scrollbarWidth: 'none' }}
      >
        <ul className="flex gap-4 mx-4">{faceElms}</ul>
        <div className="h-[200px]"></div>
      </div>
    </SidebarSection>
  )
}

export default MediaSidebarPeople
