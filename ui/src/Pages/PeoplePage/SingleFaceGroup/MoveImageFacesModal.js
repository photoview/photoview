import { gql, useLazyQuery, useMutation } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useEffect, useState } from 'react'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import SelectFaceGroupTable from './SelectFaceGroupTable'
import SelectImageFacesTable from './SelectImageFacesTable'
import { MY_FACES_QUERY } from '../PeoplePage'

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

const MoveImageFacesModal = ({ open, setOpen, faceGroup }) => {
  const [selectedImageFaces, setSelectedImageFaces] = useState([])
  const [selectedFaceGroup, setSelectedFaceGroup] = useState(null)
  const [imagesSelected, setImagesSelected] = useState(false)
  let history = useHistory()

  const [moveImageFacesMutation] = useMutation(MOVE_IMAGE_FACES_MUTATION, {
    variables: {},
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
  })

  const [loadFaceGroups, { data: faceGroupsData }] = useLazyQuery(
    MY_FACES_QUERY
  )

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

  const imageFaces = faceGroup?.imageFaces ?? []

  let table = null
  if (!imagesSelected) {
    table = (
      <SelectImageFacesTable
        imageFaces={imageFaces}
        selectedImageFaces={selectedImageFaces}
        setSelectedImageFaces={setSelectedImageFaces}
        title="Select images to move"
      />
    )
  } else {
    if (faceGroupsData) {
      const filteredFaceGroups = faceGroupsData.myFaceGroups.filter(
        x => x != faceGroup
      )
      table = (
        <SelectFaceGroupTable
          title="Select destination face group"
          faceGroups={filteredFaceGroups}
          selectedFaceGroup={selectedFaceGroup}
          setSelectedFaceGroup={setSelectedFaceGroup}
        />
      )
    } else {
      table = <div>Loading...</div>
    }
  }

  let positiveButton = null
  if (!imagesSelected) {
    positiveButton = (
      <Button
        disabled={selectedImageFaces.length == 0}
        content="Next"
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
        content="Move image faces"
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
      <Modal.Header>Move Image Faces</Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>Move selected images of this face group to another face group</p>
          {table}
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
        {positiveButton}
      </Modal.Actions>
    </Modal>
  )
}

MoveImageFacesModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  faceGroup: PropTypes.object,
}

export default MoveImageFacesModal
