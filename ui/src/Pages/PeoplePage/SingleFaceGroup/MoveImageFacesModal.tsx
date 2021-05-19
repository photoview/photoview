import { gql, useLazyQuery, useMutation } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import SelectFaceGroupTable from './SelectFaceGroupTable'
import SelectImageFacesTable from './SelectImageFacesTable'
import { MY_FACES_QUERY } from '../PeoplePage'
import {
  singleFaceGroup_faceGroup,
  singleFaceGroup_faceGroup_imageFaces,
} from './__generated__/singleFaceGroup'
import {
  myFaces,
  myFacesVariables,
  myFaces_myFaceGroups,
  myFaces_myFaceGroups_imageFaces,
} from '../__generated__/myFaces'
import { isNil } from '../../../helpers/utils'
import {
  moveImageFaces,
  moveImageFacesVariables,
} from './__generated__/moveImageFaces'
import { useTranslation } from 'react-i18next'

const MOVE_IMAGE_FACES_MUTATION = gql`
  mutation moveImageFaces($faceIDs: [ID!]!, $destFaceGroupID: ID!) {
    moveImageFaces(
      imageFaceIDs: $faceIDs
      destinationFaceGroupID: $destFaceGroupID
    ) {
      id
      imageFaces {
        id
      }
    }
  }
`

type MoveImageFacesModalProps = {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
  faceGroup: singleFaceGroup_faceGroup
}

const MoveImageFacesModal = ({
  open,
  setOpen,
  faceGroup,
}: MoveImageFacesModalProps) => {
  const { t } = useTranslation()

  const [selectedImageFaces, setSelectedImageFaces] = useState<
    (singleFaceGroup_faceGroup_imageFaces | myFaces_myFaceGroups_imageFaces)[]
  >([])
  const [selectedFaceGroup, setSelectedFaceGroup] =
    useState<myFaces_myFaceGroups | singleFaceGroup_faceGroup | null>(null)
  const [imagesSelected, setImagesSelected] = useState(false)
  const history = useHistory()

  const [moveImageFacesMutation] = useMutation<
    moveImageFaces,
    moveImageFacesVariables
  >(MOVE_IMAGE_FACES_MUTATION, {
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
  })

  const [loadFaceGroups, { data: faceGroupsData }] =
    useLazyQuery<myFaces, myFacesVariables>(MY_FACES_QUERY)

  useEffect(() => {
    if (imagesSelected) {
      loadFaceGroups()
    }
  }, [imagesSelected])

  useEffect(() => {
    if (!open) {
      setImagesSelected(false)
      setSelectedImageFaces([])
      setSelectedFaceGroup(null)
    }
  }, [open])

  if (open == false) return null

  const moveImageFaces = () => {
    const faceIDs = selectedImageFaces.map(face => face.id)

    if (isNil(selectedFaceGroup)) {
      throw new Error('Expected selectedFaceGroup not to be null')
    }

    moveImageFacesMutation({
      variables: {
        faceIDs,
        destFaceGroupID: selectedFaceGroup.id,
      },
    }).then(() => {
      setOpen(false)
      history.push(`/people/${selectedFaceGroup.id}`)
    })
  }

  const imageFaces = faceGroup.imageFaces

  let table = null
  if (!imagesSelected) {
    table = (
      <SelectImageFacesTable
        imageFaces={imageFaces}
        selectedImageFaces={selectedImageFaces}
        setSelectedImageFaces={setSelectedImageFaces}
        title={t(
          'people_page.modal.move_image_faces.image_select_table.title',
          'Select images to move'
        )}
      />
    )
  } else {
    if (faceGroupsData && faceGroup) {
      const filteredFaceGroups = faceGroupsData.myFaceGroups.filter(
        x => x.id != faceGroup.id
      )
      table = (
        <SelectFaceGroupTable
          title={t(
            'people_page.modal.move_image_faces.destination_face_group_table.title',
            'Select destination face group'
          )}
          faceGroups={filteredFaceGroups}
          selectedFaceGroup={selectedFaceGroup}
          setSelectedFaceGroup={setSelectedFaceGroup}
        />
      )
    } else {
      table = <div>{t('general.loading.default', 'Loading...')}</div>
    }
  }

  let positiveButton = null
  if (!imagesSelected) {
    positiveButton = (
      <Button
        disabled={selectedImageFaces.length == 0}
        content={t(
          'people_page.modal.move_image_faces.image_select_table.next_action',
          'Next'
        )}
        labelPosition="right"
        icon="arrow right"
        onClick={() => setImagesSelected(true)}
        positive
      />
    )
  } else {
    positiveButton = (
      <Button
        disabled={!selectedFaceGroup}
        content={t(
          'people_page.modal.move_image_faces.destination_face_group_table.move_action',
          'Move image faces'
        )}
        labelPosition="right"
        icon="checkmark"
        onClick={() => moveImageFaces()}
        positive
      />
    )
  }

  return (
    <Modal
      onClose={() => setOpen(false)}
      onOpen={() => setOpen(true)}
      open={open}
    >
      <Modal.Header>
        {t('people_page.modal.move_image_faces.title', 'Move Image Faces')}
      </Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>
            {t(
              'people_page.modal.move_image_faces.description',
              'Move selected images of this face group to another face group'
            )}
          </p>
          {table}
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>
          {t('general.action.cancel', 'Cancel')}
        </Button>
        {positiveButton}
      </Modal.Actions>
    </Modal>
  )
}

export default MoveImageFacesModal
