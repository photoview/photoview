import React, { useState } from 'react'
import PropTypes from 'prop-types'
import {
  Modal,
  Button,
  Header,
  Table,
  Input,
  Pagination,
} from 'semantic-ui-react'
import FaceCircleImage from '../FaceCircleImage'
import { gql, useMutation, useQuery } from '@apollo/client'
import { MY_FACES_QUERY } from '../PeoplePage'
import styled from 'styled-components'
import { Redirect } from 'react-router-dom'

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

const FaceCircleWrapper = styled.div`
  display: inline-block;
  border-radius: 50%;
  border: 2px solid
    ${({ $selected }) => ($selected ? `#2185c9` : 'rgba(255,255,255,0)')};
`

const FaceGroupRowStyled = styled(Table.Row)``

const FaceGroupCell = styled(Table.Cell)`
  display: flex;
  align-items: center;
`

const RowLabel = styled.span`
  ${({ $selected }) => $selected && `font-weight: bold;`}
  margin-left: 12px;
`

const FaceGroupRow = ({ faceGroup, faceSelected, setFaceSelected }) => {
  return (
    <FaceGroupRowStyled
      $selected={faceSelected}
      key={faceGroup.id}
      onClick={setFaceSelected}
    >
      <FaceGroupCell>
        <FaceCircleWrapper $selected={faceSelected}>
          <FaceCircleImage imageFace={faceGroup.imageFaces[0]} size="50px" />
        </FaceCircleWrapper>
        <RowLabel $selected={faceSelected}>{faceGroup.label}</RowLabel>
      </FaceGroupCell>
    </FaceGroupRowStyled>
  )
}

FaceGroupRow.propTypes = {
  faceGroup: PropTypes.object.isRequired,
  faceSelected: PropTypes.bool.isRequired,
  setFaceSelected: PropTypes.func.isRequired,
}

const MergeFaceGroupsModal = ({ open, setOpen, sourceFaceGroup }) => {
  const [page, setPage] = useState(0)
  const [selectedRow, setSelectedRow] = useState(null)
  const [mergedFaceGroup, setMergedFaceGroup] = useState(false)
  const PAGE_SIZE = 8
  const { data } = useQuery(MY_FACES_QUERY)
  const [combineFacesMutation] = useMutation(COMBINE_FACES_MUTATION, {
    variables: {
      srcID: sourceFaceGroup.id,
    },
    refetchQueries: [
      {
        query: MY_FACES_QUERY,
      },
    ],
  })

  const mergeFaceGroups = () => {
    const destFaceGroup = data.myFaceGroups.filter(
      x => x.id != sourceFaceGroup.id
    )[selectedRow]

    combineFacesMutation({
      variables: {
        destID: destFaceGroup.id,
      },
      onCompleted() {
        setMergedFaceGroup(destFaceGroup.id)
      },
    })
  }

  if (mergedFaceGroup) {
    return <Redirect to={`/people/${mergedFaceGroup}`} />
  }

  const rows = data?.myFaceGroups
    .filter(x => x.id != sourceFaceGroup.id)
    .filter((_, i) => i >= page * PAGE_SIZE && i < (page + 1) * PAGE_SIZE)
    .map((face, i) => (
      <FaceGroupRow
        key={face.id}
        faceGroup={face}
        faceSelected={selectedRow == i + page * PAGE_SIZE}
        setFaceSelected={() => setSelectedRow(i + page * PAGE_SIZE)}
      />
    ))

  return (
    <Modal
      onClose={() => setOpen(false)}
      onOpen={() => setOpen(true)}
      open={open}
    >
      <Modal.Header>Merge Face Groups</Modal.Header>
      <Modal.Content>
        <Modal.Description>
          <Header>Select the destination face below</Header>
          <Table selectable>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>Face group</Table.HeaderCell>
              </Table.Row>
              <Table.Row>
                <Table.HeaderCell>
                  <Input icon="search" placeholder="Search faces..." fluid />
                </Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>{rows}</Table.Body>
            <Table.Footer>
              <Table.Row>
                <Table.HeaderCell>
                  <Pagination
                    floated="right"
                    firstItem={null}
                    lastItem={null}
                    // nextItem={null}
                    // prevItem={null}
                    activePage={page + 1}
                    totalPages={data?.myFaceGroups.length / PAGE_SIZE}
                    onPageChange={(_, { activePage }) => {
                      setPage(Math.ceil(activePage) - 1)
                    }}
                  />
                </Table.HeaderCell>
              </Table.Row>
            </Table.Footer>
          </Table>
        </Modal.Description>
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
        <Button
          disabled={selectedRow == null}
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
  sourceFaceGroup: PropTypes.object.isRequired,
}

export default MergeFaceGroupsModal
