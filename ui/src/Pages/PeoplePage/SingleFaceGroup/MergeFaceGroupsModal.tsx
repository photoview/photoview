import { gql, useMutation, useQuery } from '@apollo/client'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import { isNil } from '../../../helpers/utils'
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
  sourceFaceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup
}

const MergeFaceGroupsModal = ({
  open,
  setOpen,
  sourceFaceGroup,
}: MergeFaceGroupsModalProps) => {
  const { t } = useTranslation()

  const [selectedFaceGroup, setSelectedFaceGroup] =
    useState<myFaces_myFaceGroups | singleFaceGroup_faceGroup | null>(null)

  const history = useHistory()
  const { data } = useQuery<myFaces, myFacesVariables>(MY_FACES_QUERY)
  const [combineFacesMutation] = useMutation<
    combineFaces,
    combineFacesVariables
  >(COMBINE_FACES_MUTATION, {
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
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
      history.push(`/people/${selectedFaceGroup.id}`)
    })
  }

  return (
    <Modal
      onClose={() => setOpen(false)}
      onOpen={() => setOpen(true)}
      open={open}
    >
      <Modal.Header>
        {t('people_page.modal.merge_face_groups.title', 'Merge Face Groups')}
      </Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>
            {t(
              'people_page.modal.merge_face_groups.description',
              'All images within this face group will be merged into the selected face group.'
            )}
          </p>
          <SelectFaceGroupTable
            title={t(
              'people_page.modal.merge_face_groups.destination_table.title',
              'Select the destination face'
            )}
            faceGroups={filteredFaceGroups}
            selectedFaceGroup={selectedFaceGroup}
            setSelectedFaceGroup={setSelectedFaceGroup}
          />
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>
          {t('general.action.cancel', 'Cancel')}
        </Button>
        <Button
          disabled={selectedFaceGroup == null}
          content={t('people_page.modal.action.merge', 'Merge')}
          labelPosition="right"
          icon="checkmark"
          onClick={() => mergeFaceGroups()}
          positive
        />
      </Modal.Actions>
    </Modal>
  )
}

export default MergeFaceGroupsModal
