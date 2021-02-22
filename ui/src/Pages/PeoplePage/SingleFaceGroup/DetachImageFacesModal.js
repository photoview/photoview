import { gql, useMutation } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useEffect, useState } from 'react'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import { MY_FACES_QUERY } from '../PeoplePage'
import SelectImageFacesTable from './SelectImageFacesTable'

const DETACH_IMAGE_FACES_MUTATION = gql`
  mutation detachImageFaces($faceIDs: [ID!]!) {
    detachImageFaces(imageFaceIDs: $faceIDs) {
      id
      label
    }
  }
`

const DetachImageFacesModal = ({ open, setOpen, faceGroup }) => {
  const [selectedImageFaces, setSelectedImageFaces] = useState([])
  let history = useHistory()

  const [detachImageFacesMutation] = useMutation(DETACH_IMAGE_FACES_MUTATION, {
    variables: {},
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
      <Modal.Header>Detach Image Faces</Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>
            Detach selected images of this face group and move them to a new
            face group
          </p>
          <SelectImageFacesTable
            imageFaces={imageFaces}
            selectedImageFaces={selectedImageFaces}
            setSelectedImageFaces={setSelectedImageFaces}
            title="Select images to detach"
          />
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
        <Button
          disabled={selectedImageFaces.length == 0}
          content="Detach image faces"
          labelPosition="right"
          icon="checkmark"
          onClick={() => detachImageFaces()}
          positive
        />
      </Modal.Actions>
    </Modal>
  )
}

DetachImageFacesModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  faceGroup: PropTypes.object,
}

export default DetachImageFacesModal
