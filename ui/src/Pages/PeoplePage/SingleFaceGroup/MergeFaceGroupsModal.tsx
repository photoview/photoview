import { gql, PureQueryOptions, useMutation, useQuery } from '@apollo/client'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { isNil } from '../../../helpers/utils'
import Modal from '../../../primitives/Modal'
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
  mutation combineFaces($destID: ID!, $srcID: ID!) {
    combineFaceGroups(
      destinationFaceGroupID: $destID
      sourceFaceGroupID: $srcID
    ) {
      id
    }
  }
`

type MergeFaceGroupsModalProps = {
  open: boolean
  setOpen(open: boolean): void
  sourceFaceGroup: {
    __typename: 'FaceGroup'
    id: string
  }
  refetchQueries: PureQueryOptions[]
}

const MergeFaceGroupsModal = ({
  open,
  setOpen,
  sourceFaceGroup,
  refetchQueries,
}: MergeFaceGroupsModalProps) => {
  const { t } = useTranslation()

  const [selectedFaceGroup, setSelectedFaceGroup] = useState<
    myFaces_myFaceGroups | singleFaceGroup_faceGroup | null
  >(null)

  const navigate = useNavigate()
  const { data } = useQuery<myFaces, myFacesVariables>(MY_FACES_QUERY)
  const [combineFacesMutation] = useMutation<
    combineFaces,
    combineFacesVariables
  >(COMBINE_FACES_MUTATION, {
    refetchQueries: refetchQueries,
  })

  if (open == false) return null

  const filteredFaceGroups =
    data?.myFaceGroups.filter(x => x.id != sourceFaceGroup?.id) ?? []

  const mergeFaceGroups = () => {
    if (isNil(selectedFaceGroup)) throw new Error('No selected face group')

    combineFacesMutation({
      variables: {
        srcID: sourceFaceGroup.id,
        destID: selectedFaceGroup.id,
      },
    }).then(() => {
      setOpen(false)
      navigate(`/people/${selectedFaceGroup.id}`)
    })
  }

  return (
    <Modal
      title={t(
        'people_page.modal.merge_face_groups.title',
        'Merge Face Groups'
      )}
      description={t(
        'people_page.modal.merge_face_groups.description',
        'All images within this face group will be merged into the selected face group.'
      )}
      actions={[
        {
          key: 'cancel',
          label: t('general.action.cancel', 'Cancel'),
          onClick: () => setOpen(false),
        },
        {
          key: 'merge',
          label: t('people_page.modal.action.merge', 'Merge'),
          onClick: () => mergeFaceGroups(),
          variant: 'positive',
        },
      ]}
      onClose={() => setOpen(false)}
      open={open}
    >
      <SelectFaceGroupTable
        title={t(
          'people_page.modal.merge_face_groups.destination_table.title',
          'Select the destination face'
        )}
        faceGroups={filteredFaceGroups}
        selectedFaceGroup={selectedFaceGroup}
        setSelectedFaceGroup={setSelectedFaceGroup}
      />
    </Modal>
  )
}

export default MergeFaceGroupsModal
