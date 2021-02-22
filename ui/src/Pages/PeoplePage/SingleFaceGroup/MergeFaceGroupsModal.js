import { gql, useMutation, useQuery } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useState } from 'react'
import { useHistory } from 'react-router-dom'
import { Button, Modal } from 'semantic-ui-react'
import { MY_FACES_QUERY } from '../PeoplePage'
import SelectFaceGroupTable from './SelectFaceGroupTable'

const COMBINE_FACES_MUTATION = gql`
  mutation($destID: ID!, $srcID: ID!) {
    combineFaceGroups(
      destinationFaceGroupID: $destID
      sourceFaceGroupID: $srcID
    ) {
      id
    }
  }
`

const MergeFaceGroupsModal = ({ open, setOpen, sourceFaceGroup }) => {
  const [selectedFaceGroup, setSelectedFaceGroup] = useState(null)

  let history = useHistory()
  const { data } = useQuery(MY_FACES_QUERY)
  const [combineFacesMutation] = useMutation(COMBINE_FACES_MUTATION, {
    variables: {
      srcID: sourceFaceGroup?.id,
    },
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
    combineFacesMutation({
      variables: {
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
      <Modal.Header>Merge Face Groups</Modal.Header>
      <Modal.Content scrolling>
        <Modal.Description>
          <p>
            All images within this face group will be merged into the selected
            face group.
          </p>
          <SelectFaceGroupTable
            title="Select the destination face"
            faceGroups={filteredFaceGroups}
            selectedFaceGroup={selectedFaceGroup}
            setSelectedFaceGroup={setSelectedFaceGroup}
          />
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
        <Button
          disabled={selectedFaceGroup == null}
          content="Merge"
          labelPosition="right"
          icon="checkmark"
          onClick={() => mergeFaceGroups()}
          positive
        />
      </Modal.Actions>
    </Modal>
  )
}

MergeFaceGroupsModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  sourceFaceGroup: PropTypes.object,
}

export default MergeFaceGroupsModal
