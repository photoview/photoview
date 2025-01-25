import { gql, PureQueryOptions, useMutation, useQuery } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { isNil } from '../../../helpers/utils'
import Modal, { ModalAction, ModalProps } from '../../../primitives/Modal'
import { MY_FACES_QUERY } from '../PeoplePage'
import {
  myFaces,
  myFacesVariables,
  myFaces_myFaceGroups,
} from '../__generated__/myFaces'
import SelectFaceGroupTable from './SelectFaceGroupTable'
import {
  combineFaces,
  combineFacesVariables,
} from './__generated__/combineFaces'
import { singleFaceGroup_faceGroup } from './__generated__/singleFaceGroup'

const COMBINE_FACES_MUTATION = gql`
  mutation combineFaces($destID: ID!, $srcIDs: [ID!]!) {
    combineFaceGroups(
      destinationFaceGroupID: $destID
      sourceFaceGroupIDs: $srcIDs
    ) {
      id
    }
  }
`

export enum MergeFaceGroupsModalState {
  Closed = 'closed',
  SelectDestination = 'select_destination',
  SelectSources = 'select_sources',
}

type MergeFaceGroupsModalProps = {
  state: MergeFaceGroupsModalState
  setState(state: MergeFaceGroupsModalState): void
  preselectedDestinationFaceGroup?: {
    __typename: 'FaceGroup'
    id: string
  }
  refetchQueries: PureQueryOptions[]
}

type StateContent = {
  props: ModalProps
  searchTitle: string
}

const MergeFaceGroupsModal = ({
  state,
  setState,
  preselectedDestinationFaceGroup,
  refetchQueries,
}: MergeFaceGroupsModalProps) => {
  const { t } = useTranslation()

  const navigate = useNavigate()
  const { data } = useQuery<myFaces, myFacesVariables>(MY_FACES_QUERY)
  const [combineFacesMutation] = useMutation<
    combineFaces,
    combineFacesVariables
  >(COMBINE_FACES_MUTATION, {
    refetchQueries: refetchQueries,
  })

  // The destination face group
  const [selectedDestinationFaceGroup, setSelectedDestinationFaceGroup] =
    useState<myFaces_myFaceGroups | singleFaceGroup_faceGroup | null>(null)

  // The set of currently selected face groups, on the modal page
  const [selectedFaceGroups, setSelectedFaceGroups] = useState<
    Set<myFaces_myFaceGroups | singleFaceGroup_faceGroup | null>
  >(new Set())

  const addSelectedFaceGroup = (
    faceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup | null
  ) => setSelectedFaceGroups(prev => new Set(prev).add(faceGroup))
  const removeSelectedFaceGroup = (
    faceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup | null
  ) => {
    setSelectedFaceGroups(prev => {
      const s = new Set(prev)
      s.delete(faceGroup)
      return s
    })
  }

  const setDestinationFaceGroup = (
    faceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup | null
  ) => {
    if (isNil(faceGroup)) {
      setSelectedFaceGroups(new Set())
      setSelectedDestinationFaceGroup(null)
      return
    }

    // Overwrite the selected face groups with a set containing only the selected group
    setSelectedFaceGroups(
      new Set<myFaces_myFaceGroups | singleFaceGroup_faceGroup | null>().add(
        faceGroup
      )
    )

    setSelectedDestinationFaceGroup(faceGroup)
  }

  // Go straight to the sources page if a destination face group is preselected, using the preselection as the destination
  useEffect(() => {
    if (isNil(preselectedDestinationFaceGroup)) return
    if (state != MergeFaceGroupsModalState.SelectDestination) return

    const destinationFaceGroup = data?.myFaceGroups.find(
      x => x.id == preselectedDestinationFaceGroup?.id
    )
    if (isNil(destinationFaceGroup)) return

    setDestinationFaceGroup(destinationFaceGroup)
  }, [state, preselectedDestinationFaceGroup])

  function handleFaceGroupToggled(newValue: myFaces_myFaceGroups | singleFaceGroup_faceGroup) {
    switch (state) {
      case MergeFaceGroupsModalState.SelectDestination:
        setSelectedDestinationFaceGroup(newValue)
        setSelectedFaceGroups(new Set([newValue]))
        break
      case MergeFaceGroupsModalState.SelectSources:
        if (selectedFaceGroups.has(newValue))
          removeSelectedFaceGroup(newValue)
        else addSelectedFaceGroup(newValue)
        break
    }
  }

  // Show all face groups on the destination page, but filter out the destination group on the source page
  const filteredFaceGroups =
  data?.myFaceGroups.filter(
    x =>
      state === MergeFaceGroupsModalState.SelectDestination ||
      x.id != selectedDestinationFaceGroup?.id
  ) ?? []

  const goNext = () => {
    if (isNil(selectedDestinationFaceGroup))
      throw new Error('No selected face group')

    setState(MergeFaceGroupsModalState.SelectSources)
    setSelectedFaceGroups(new Set())
  }

  const mergeFaceGroups = () => {
    if (isNil(selectedDestinationFaceGroup))
      throw new Error('No selected destination face group')

    const sourceGroupIDs: string[] = [...selectedFaceGroups].filter(fc => fc !== null).map(fc => fc.id)

    if(sourceGroupIDs.length < 1)
      throw new Error('No selected source face groups')

    combineFacesMutation({
      variables: {
        srcIDs: sourceGroupIDs,
        destID: selectedDestinationFaceGroup.id,
      },
    }).then(() => {
      setState(MergeFaceGroupsModalState.Closed)
      navigate(`/people/${selectedDestinationFaceGroup.id}`)
    })
  }

  const closeModal = () => {
    setState(MergeFaceGroupsModalState.Closed)
  }

  const isOpen: boolean = state !== MergeFaceGroupsModalState.Closed

  const cancelAction: ModalAction = {
    key: 'cancel',
    label: t('general.action.cancel', 'Cancel'),
    onClick: () => setState(MergeFaceGroupsModalState.Closed),
  }

  const nextAction: ModalAction = {
    key: 'next',
    label: t('people_page.modal.action.next', 'Next'),
    onClick: () => goNext(),
    variant: 'positive',
  }

  const mergeAction: ModalAction = {
    key: 'merge',
    label: t('people_page.modal.action.merge', 'Merge'),
    onClick: () => mergeFaceGroups(),
    variant: 'positive',
  }

  const modalTitle: string = t(
    'people_page.modal.merge_face_groups.title',
    'Merge Face Groups'
  )

  const selectDestinationProps: StateContent = {
    props: {
      title: modalTitle,
      description: t(
        'people_page.modal.merge_face_groups.destination_description',
        'Select the face group that other groups should be merged into.'
      ),
      actions: [cancelAction, nextAction],
      onClose: closeModal,
      open: isOpen,
    },
    searchTitle: t(
      'people_page.modal.merge_face_groups.destination_table.title',
      'Select the destination face'
    ),
  }

  const selectSourcesProps: StateContent = {
    props: {
      title: modalTitle,
      description: t(
        'people_page.modal.merge_face_groups.sources_description',
        'Select all face groups that will be merged into the destination group.'
      ),
      actions: [cancelAction, mergeAction],
      onClose: closeModal,
      open: isOpen,
    },
    searchTitle:
      t(
        'people_page.modal.merge_face_groups.sources_table.title',
        'Select one or more source faces to merge into:'
      ) +
      ` ${
        selectedDestinationFaceGroup?.label ??
        t('people_page.face_group.unlabeled', 'Unlabeled') ??
        'Unlabeled'
      }`,
  }

  const modalContent: StateContent =
    state === MergeFaceGroupsModalState.SelectDestination
      ? selectDestinationProps
      : selectSourcesProps

  return (
    <Modal {...modalContent.props}>
      <SelectFaceGroupTable
        title={modalContent.searchTitle}
        frozen={state === MergeFaceGroupsModalState.SelectDestination && preselectedDestinationFaceGroup !== undefined}
        faceGroups={filteredFaceGroups}
        selectedFaceGroups={selectedFaceGroups}
        toggleSelectedFaceGroup={(face) => { if(face === null) return; handleFaceGroupToggled(face) }}
      />
    </Modal>
  )
}

export default MergeFaceGroupsModal
