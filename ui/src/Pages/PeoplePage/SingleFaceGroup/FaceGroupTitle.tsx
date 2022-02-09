import { useMutation } from '@apollo/client'
import React, {
  useState,
  useEffect,
  createRef,
  KeyboardEventHandler,
} from 'react'
import { useTranslation } from 'react-i18next'
import { isNil } from '../../../helpers/utils'
import { Button, TextField } from '../../../primitives/form/Input'
import { MY_FACES_QUERY, SET_GROUP_LABEL_MUTATION } from '../PeoplePage'
import {
  setGroupLabel,
  setGroupLabelVariables,
} from '../__generated__/setGroupLabel'
import DetachImageFacesModal from './DetachImageFacesModal'
import MergeFaceGroupsModal from './MergeFaceGroupsModal'
import MoveImageFacesModal from './MoveImageFacesModal'
import { singleFaceGroup_faceGroup } from './__generated__/singleFaceGroup'

type FaceGroupTitleProps = {
  faceGroup?: singleFaceGroup_faceGroup
}

const FaceGroupTitle = ({ faceGroup }: FaceGroupTitleProps) => {
  const { t } = useTranslation()

  const [editLabel, setEditLabel] = useState(false)
  const [inputValue, setInputValue] = useState(faceGroup?.label ?? '')
  const inputRef = createRef<HTMLInputElement>()
  const [mergeModalOpen, setMergeModalOpen] = useState(false)
  const [moveModalOpen, setMoveModalOpen] = useState(false)
  const [detachModalOpen, setDetachModalOpen] = useState(false)

  const [setGroupLabel, { loading: setLabelLoading }] = useMutation<
    setGroupLabel,
    setGroupLabelVariables
  >(SET_GROUP_LABEL_MUTATION)

  const resetLabel = () => {
    setInputValue(faceGroup?.label ?? '')
    setEditLabel(false)
  }

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus()
    }
  }, [inputRef])

  useEffect(() => {
    if (!setLabelLoading) {
      resetLabel()
    }
  }, [setLabelLoading])

  const onKeyDown: KeyboardEventHandler<HTMLInputElement> = e => {
    if (e.key == 'Escape') {
      resetLabel()
      return
    }
  }

  let title
  if (!editLabel) {
    title = (
      <>
        <h1
          className={`text-2xl font-semibold ${
            faceGroup?.label ? '' : 'text-gray-600 dark:text-gray-400'
          }`}
        >
          {faceGroup?.label ??
            t('people_page.face_group.unlabeled_person', 'Unlabeled person')}
        </h1>
      </>
    )
  } else {
    title = (
      <>
        <TextField
          loading={setLabelLoading}
          ref={inputRef}
          placeholder={t('people_page.face_group.label_placeholder', 'Label')}
          action={() => {
            if (isNil(faceGroup))
              throw new Error('Expected faceGroup to be defined')

            setGroupLabel({
              variables: {
                groupID: faceGroup.id,
                label: inputValue ? inputValue : null,
              },
            })
          }}
          value={inputValue}
          onKeyDown={onKeyDown}
          onChange={e => setInputValue(e.target.value)}
          onBlur={() => {
            resetLabel()
          }}
        />
      </>
    )
  }

  let modals = null
  if (faceGroup) {
    modals = (
      <>
        <MergeFaceGroupsModal
          open={mergeModalOpen}
          setOpen={setMergeModalOpen}
          sourceFaceGroup={faceGroup}
          refetchQueries={[
            {
              query: MY_FACES_QUERY,
            },
          ]}
        />
        <MoveImageFacesModal
          open={moveModalOpen}
          setOpen={setMoveModalOpen}
          faceGroup={faceGroup}
        />
        <DetachImageFacesModal
          open={detachModalOpen}
          setOpen={setDetachModalOpen}
          faceGroup={faceGroup}
        />
      </>
    )
  }

  return (
    <>
      <div>
        <div className="mb-2">{title}</div>
        <ul className="flex gap-2 flex-wrap mb-6">
          <li>
            <Button onClick={() => setEditLabel(true)}>
              {t('people_page.action_label.change_label', 'Change label')}
            </Button>
          </li>
          <li>
            <Button onClick={() => setMergeModalOpen(true)}>
              {t('people_page.action_label.merge_people', 'Merge people')}
            </Button>
          </li>
          <li>
            <Button onClick={() => setDetachModalOpen(true)}>
              {t('people_page.action_label.detach_images', 'Detach images')}
            </Button>
          </li>
          <li>
            <Button onClick={() => setMoveModalOpen(true)}>
              {t('people_page.action_label.move_faces', 'Move faces')}
            </Button>
          </li>
        </ul>
      </div>
      {modals}
    </>
  )
}

export default FaceGroupTitle
