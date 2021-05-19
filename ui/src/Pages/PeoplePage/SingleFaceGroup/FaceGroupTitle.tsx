import { useMutation } from '@apollo/client'
import React, { useState, useEffect, createRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Dropdown, Input } from 'semantic-ui-react'
import styled from 'styled-components'
import { isNil } from '../../../helpers/utils'
import { SET_GROUP_LABEL_MUTATION } from '../PeoplePage'
import {
  setGroupLabel,
  setGroupLabelVariables,
} from '../__generated__/setGroupLabel'
import DetachImageFacesModal from './DetachImageFacesModal'
import MergeFaceGroupsModal from './MergeFaceGroupsModal'
import MoveImageFacesModal from './MoveImageFacesModal'
import { singleFaceGroup_faceGroup } from './__generated__/singleFaceGroup'

const TitleWrapper = styled.div`
  min-height: 3.5em;
`

const TitleLabel = styled.h1<{ labeled: boolean }>`
  display: inline-block;
  color: ${({ labeled }) => (labeled ? 'black' : '#888')};
  margin-right: 12px;
`

const TitleDropdown = styled(Dropdown)`
  vertical-align: middle;
  margin-top: -10px;
  color: #888;

  &:hover {
    color: #1e70bf;
  }
`

type FaceGroupTitleProps = {
  faceGroup?: singleFaceGroup_faceGroup
}

const FaceGroupTitle = ({ faceGroup }: FaceGroupTitleProps) => {
  const { t } = useTranslation()

  const [editLabel, setEditLabel] = useState(false)
  const [inputValue, setInputValue] = useState(faceGroup?.label ?? '')
  const inputRef = createRef<Input>()
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

  const onKeyUp = (e: KeyboardEvent & React.ChangeEvent<HTMLInputElement>) => {
    if (isNil(faceGroup)) throw new Error('Expected faceGroup to be defined')

    if (e.key == 'Escape') {
      resetLabel()
      return
    }

    if (e.key == 'Enter') {
      setGroupLabel({
        variables: {
          groupID: faceGroup.id,
          label: e.target.value == '' ? null : e.target.value,
        },
      })
      return
    }
  }

  let title
  if (!editLabel) {
    title = (
      <TitleWrapper>
        <TitleLabel labeled={!!faceGroup?.label}>
          {faceGroup?.label ??
            t('people_page.face_group.unlabeled_person', 'Unlabeled person')}
        </TitleLabel>
        <TitleDropdown
          icon={{
            name: 'settings',
            size: 'large',
          }}
        >
          <Dropdown.Menu>
            <Dropdown.Item
              icon="pencil"
              text={
                faceGroup?.label
                  ? t(
                      'people_page.face_group.action.change_label',
                      'Change Label'
                    )
                  : t('people_page.face_group.action.add_label', 'Add Label')
              }
              onClick={() => setEditLabel(true)}
            />
            <Dropdown.Item
              icon="object group"
              text={t('people_page.face_group.action.merge_face', 'Merge Face')}
              onClick={() => setMergeModalOpen(true)}
            />
            <Dropdown.Item
              icon="object ungroup"
              text={t(
                'people_page.face_group.action.detach_face',
                'Detach Face'
              )}
              onClick={() => setDetachModalOpen(true)}
            />
            <Dropdown.Item
              icon="clone"
              text={t('people_page.face_group.action.move_faces', 'Move Faces')}
              onClick={() => setMoveModalOpen(true)}
            />
          </Dropdown.Menu>
        </TitleDropdown>
      </TitleWrapper>
    )
  } else {
    title = (
      <TitleWrapper>
        <Input
          loading={setLabelLoading}
          ref={inputRef}
          placeholder={t('people_page.face_group.label_placeholder', 'Label')}
          icon="arrow right"
          value={inputValue}
          onKeyUp={onKeyUp}
          onChange={e => setInputValue(e.target.value)}
          onBlur={() => {
            resetLabel()
          }}
        />
      </TitleWrapper>
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
      {title}
      {modals}
    </>
  )
}

export default FaceGroupTitle
