import { BaseMutationOptions, gql, useMutation } from '@apollo/client'
import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { isNil } from '../../../helpers/utils'
import Modal from '../../../primitives/Modal'
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

export const useDetachImageFaces = (
  mutationOptions: BaseMutationOptions<
    detachImageFaces,
    detachImageFacesVariables
  >
) => {
  const [detachImageFacesMutation] = useMutation<
    detachImageFaces,
    detachImageFacesVariables
  >(DETACH_IMAGE_FACES_MUTATION, mutationOptions)

  return async (
    selectedImageFaces: (
      | myFaces_myFaceGroups_imageFaces
      | singleFaceGroup_faceGroup_imageFaces
    )[]
  ) => {
    const faceIDs = selectedImageFaces.map(face => face.id)

    const result = await detachImageFacesMutation({
      variables: {
        faceIDs,
      },
    })

    return result
  }
}

type DetachImageFacesModalProps = {
  open: boolean
  setOpen(open: boolean): void
  faceGroup: myFaces_myFaceGroups | singleFaceGroup_faceGroup
  selectedImageFaces?: (
    | myFaces_myFaceGroups_imageFaces
    | singleFaceGroup_faceGroup_imageFaces
  )[]
}

const DetachImageFacesModal = ({
  open,
  setOpen,
  faceGroup,
  selectedImageFaces: selectedImageFacesProp,
}: DetachImageFacesModalProps) => {
  const { t } = useTranslation()

  const [selectedImageFaces, setSelectedImageFaces] = useState<
    (myFaces_myFaceGroups_imageFaces | singleFaceGroup_faceGroup_imageFaces)[]
  >([])
  const navigate = useNavigate()

  const detachImageFacesMutation = useDetachImageFaces({
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
  })

  const detachImageFaces = () => {
    detachImageFacesMutation(selectedImageFaces).then(({ data }) => {
      if (isNil(data)) throw new Error('Expected data not to be null')
      setOpen(false)
      navigate(`/people/${data.detachImageFaces.id}`)
    })
  }

  useEffect(() => {
    if (isNil(selectedImageFacesProp)) return
    setSelectedImageFaces(selectedImageFacesProp)
  }, [selectedImageFacesProp])

  useEffect(() => {
    if (!open) {
      setSelectedImageFaces([])
    }
  }, [open])

  if (open == false) return null

  const imageFaces = faceGroup?.imageFaces ?? []

  return (
    <Modal
      title={t(
        'people_page.modal.detach_image_faces.title',
        'Detach Image Faces'
      )}
      description={t(
        'people_page.modal.detach_image_faces.description',
        'Detach selected images of this face group and move them to a new face groups'
      )}
      actions={[
        {
          key: 'cancel',
          label: t('general.action.cancel', 'Cancel'),
          onClick: () => setOpen(false),
        },
        {
          key: 'detach',
          label: t(
            'people_page.modal.detach_image_faces.action.detach',
            'Detach image faces'
          ),
          variant: 'positive',
          onClick: () => detachImageFaces(),
        },
      ]}
      onClose={() => setOpen(false)}
      open={open}
    >
      <SelectImageFacesTable
        imageFaces={imageFaces}
        selectedImageFaces={selectedImageFaces}
        setSelectedImageFaces={setSelectedImageFaces}
        title={t(
          'people_page.modal.detach_image_faces.action.select_images',
          'Select images to detach'
        )}
      />
    </Modal>
  )
}

export default DetachImageFacesModal
