import { gql, useMutation } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import { isNil } from '../../../helpers/utils'
import { MY_FACES_QUERY } from '../PeoplePage'
import {
  myFaces_myFaceGroups,
  myFaces_myFaceGroups_imageFaces,
} from '../__generated__/myFaces'
import SelectImageFacesTable from './SelectImageFacesTable'
import {
  detachImageFaces,
  detachImageFacesVariables,
} from './__generated__/detachImageFaces'
import {
  singleFaceGroup_faceGroup,
  singleFaceGroup_faceGroup_imageFaces,
} from './__generated__/singleFaceGroup'

const DETACH_IMAGE_FACES_MUTATION = gql`
  mutation detachImageFaces($faceIDs: [ID!]!) {
    detachImageFaces(imageFaceIDs: $faceIDs) {
      id
      label
    }
  }
`

type DetachImageFacesModalProps = {
  open: boolean
  setOpen(open: boolean): void
  faceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup
}

const DetachImageFacesModal = ({
  open,
  setOpen,
  faceGroup,
}: DetachImageFacesModalProps) => {
  const { t } = useTranslation()

  const [selectedImageFaces, setSelectedImageFaces] = useState<
    (myFaces_myFaceGroups_imageFaces | singleFaceGroup_faceGroup_imageFaces)[]
  >([])
  const history = useHistory()

  const [detachImageFacesMutation] = useMutation<
    detachImageFaces,
    detachImageFacesVariables
  >(DETACH_IMAGE_FACES_MUTATION, {
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
  })

  useEffect(() => {
    if (!open) {
      setSelectedImageFaces([])
    }
  }, [open])

  if (open == false) return null

  const detachImageFaces = () => {
    const faceIDs = selectedImageFaces.map(face => face.id)

    detachImageFacesMutation({
      variables: {
        faceIDs,
      },
    }).then(({ data }) => {
      if (isNil(data)) throw new Error('Expected data not to be null')
      setOpen(false)
      history.push(`/people/${data.detachImageFaces.id}`)
    })
  }

  const imageFaces = faceGroup?.imageFaces ?? []

  return (
    <Modal
      onClose={() => setOpen(false)}
      onOpen={() => setOpen(true)}
      open={open}
    >
      <Modal.Header>
        {t('people_page.modal.detach_image_faces.title', 'Detach Image Faces')}
      </Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>
            {t(
              'people_page.modal.detach_image_faces.description',
              'Detach selected images of this face group and move them to a new face groups'
            )}
          </p>
          <SelectImageFacesTable
            imageFaces={imageFaces}
            selectedImageFaces={selectedImageFaces}
            setSelectedImageFaces={setSelectedImageFaces}
            title={t(
              'people_page.modal.detach_image_faces.action.select_images',
              'Select images to detach'
            )}
          />
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
        <Button
          disabled={selectedImageFaces.length == 0}
          content={t(
            'people_page.modal.detach_image_faces.action.detach',
            'Detach image faces'
          )}
          labelPosition="right"
          icon="checkmark"
          onClick={() => detachImageFaces()}
          positive
        />
      </Modal.Actions>
    </Modal>
  )
}

export default DetachImageFacesModal
