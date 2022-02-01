import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link, useNavigate } from 'react-router-dom'
import FaceCircleImage from '../../../Pages/PeoplePage/FaceCircleImage'
import { SidebarSection, SidebarSectionTitle } from '../SidebarComponents'
import { MediaSidebarMedia, SIDEBAR_MEDIA_QUERY } from './MediaSidebar'
import { sidebarMediaQuery_media_faces } from './__generated__/sidebarMediaQuery'

import { ReactComponent as PeopleDotsIcon } from './icons/peopleDotsIcon.svg'
import { Menu } from '@headlessui/react'
import { Button } from '../../../primitives/form/Input'
import { ArrowPopoverPanel } from '../Sharing'
import { isNil, tailwindClassNames } from '../../../helpers/utils'
import MergeFaceGroupsModal from '../../../Pages/PeoplePage/SingleFaceGroup/MergeFaceGroupsModal'
import { useDetachImageFaces } from '../../../Pages/PeoplePage/SingleFaceGroup/DetachImageFacesModal'
import MoveImageFacesModal from '../../../Pages/PeoplePage/SingleFaceGroup/MoveImageFacesModal'
import { FaceDetails } from '../../../Pages/PeoplePage/PeoplePage'

type PersonMoreMenuItemProps = {
  label: string
  className?: string
  onClick(): void
}

const PersonMoreMenuItem = ({
  label,
  className,
  onClick,
}: PersonMoreMenuItemProps) => {
  return (
    <Menu.Item>
      {({ active }) => (
        <button
          onClick={onClick}
          className={tailwindClassNames(
            `whitespace-normal w-full block py-1 cursor-pointer ${
              active ? 'bg-gray-50 text-black' : 'text-gray-700'
            }`,
            className
          )}
        >
          {label}
        </button>
      )}
    </Menu.Item>
  )
}

type PersonMoreMenuProps = {
  face: sidebarMediaQuery_media_faces
  setChangeLabel: React.Dispatch<React.SetStateAction<boolean>>
  className?: string
  menuFlipped: boolean
}

const PersonMoreMenu = ({
  menuFlipped,
  face,
  setChangeLabel,
  className,
}: PersonMoreMenuProps) => {
  const { t } = useTranslation()

  const [mergeModalOpen, setMergeModalOpen] = useState(false)
  const [moveModalOpen, setMoveModalOpen] = useState(false)

  const refetchQueries = [
    {
      query: SIDEBAR_MEDIA_QUERY,
      variables: {
        id: face.media.id,
      },
    },
  ]

  const navigate = useNavigate()
  const detachImageFaceMutation = useDetachImageFaces({
    refetchQueries,
  })

  const modals = (
    <>
      <MergeFaceGroupsModal
        sourceFaceGroup={face.faceGroup}
        open={mergeModalOpen}
        setOpen={setMergeModalOpen}
        refetchQueries={refetchQueries}
      />
      <MoveImageFacesModal
        faceGroup={{ imageFaces: [], ...face.faceGroup }}
        open={moveModalOpen}
        setOpen={setMoveModalOpen}
        preselectedImageFaces={[face]}
      />
    </>
  )

  const detachImageFace = () => {
    if (
      !confirm(
        t(
          'sidebar.people.confirm_image_detach',
          'Are you sure you want to detach this image?'
        )
      )
    )
      return
    detachImageFaceMutation([face]).then(({ data }) => {
      if (isNil(data)) throw new Error('Expected data not to be null')
      navigate(`/people/${data.detachImageFaces.id}`)
    })
  }

  return (
    <>
      <Menu
        as="div"
        className={tailwindClassNames('relative inline-block', className)}
      >
        <Menu.Button as={Button} className="px-1.5 py-1.5 align-middle ml-1">
          <PeopleDotsIcon className="text-gray-500" />
        </Menu.Button>
        <Menu.Items className="">
          <ArrowPopoverPanel width={120} flipped={menuFlipped}>
            <PersonMoreMenuItem
              onClick={() => setChangeLabel(true)}
              className="border-b"
              label={t('people_page.action_label.change_label', 'Change label')}
            />
            <PersonMoreMenuItem
              onClick={() => setMergeModalOpen(true)}
              className="border-b"
              label={t('sidebar.people.action_label.merge_face', 'Merge face')}
            />
            <PersonMoreMenuItem
              onClick={() => detachImageFace()}
              className="border-b"
              label={t(
                'sidebar.people.action_label.detach_image',
                'Detach image'
              )}
            />
            <PersonMoreMenuItem
              onClick={() => setMoveModalOpen(true)}
              label={t('sidebar.people.action_label.move_face', 'Move face')}
            />
          </ArrowPopoverPanel>
        </Menu.Items>
      </Menu>
      {modals}
    </>
  )
}

type MediaSidebarFaceProps = {
  face: sidebarMediaQuery_media_faces
  menuFlipped: boolean
}

const MediaSidebarPerson = ({ face, menuFlipped }: MediaSidebarFaceProps) => {
  const [changeLabel, setChangeLabel] = useState(false)

  return (
    <li className="inline-block">
      <Link to={`/people/${face.faceGroup.id}`}>
        <FaceCircleImage imageFace={face} selectable={true} size="92px" />
      </Link>
      <div className="mt-1 whitespace-nowrap">
        <FaceDetails
          className="text-sm max-w-[80px] align-middle"
          textFieldClassName="w-[100px]"
          group={face.faceGroup}
          editLabel={changeLabel}
          setEditLabel={setChangeLabel}
        />
        {!changeLabel && (
          <PersonMoreMenu
            menuFlipped={menuFlipped}
            className="pl-0.5"
            face={face}
            setChangeLabel={setChangeLabel}
          />
        )}
      </div>
    </li>
  )
}

type MediaSidebarFacesProps = {
  media: MediaSidebarMedia
}

const MediaSidebarPeople = ({ media }: MediaSidebarFacesProps) => {
  const { t } = useTranslation()

  const faceElms = (media.faces ?? []).map((face, i) => (
    <MediaSidebarPerson key={face.id} face={face} menuFlipped={i == 0} />
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
