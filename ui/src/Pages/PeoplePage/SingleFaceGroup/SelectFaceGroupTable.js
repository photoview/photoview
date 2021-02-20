import PropTypes from 'prop-types'
import React, { useState, useEffect } from 'react'
import { Input, Pagination, Table } from 'semantic-ui-react'
import styled from 'styled-components'
import FaceCircleImage from '../FaceCircleImage'

const FaceCircleWrapper = styled.div`
  display: inline-block;
  border-radius: 50%;
  border: 2px solid
    ${({ $selected }) => ($selected ? `#2185c9` : 'rgba(255,255,255,0)')};
`

const FlexCell = styled(Table.Cell)`
  display: flex;
  align-items: center;
`

export const RowLabel = styled.span`
  ${({ $selected }) => $selected && `font-weight: bold;`}
  margin-left: 12px;
`

const FaceGroupRow = ({ faceGroup, faceSelected, setFaceSelected }) => {
  return (
    <Table.Row key={faceGroup.id} onClick={setFaceSelected}>
      <FlexCell>
        <FaceCircleWrapper $selected={faceSelected}>
          <FaceCircleImage imageFace={faceGroup.imageFaces[0]} size="50px" />
        </FaceCircleWrapper>
        <RowLabel $selected={faceSelected}>{faceGroup.label}</RowLabel>
      </FlexCell>
    </Table.Row>
  )
}

FaceGroupRow.propTypes = {
  faceGroup: PropTypes.object.isRequired,
  faceSelected: PropTypes.bool.isRequired,
  setFaceSelected: PropTypes.func.isRequired,
}

const SelectFaceGroupTable = ({
  faceGroups,
  selectedFaceGroup,
  setSelectedFaceGroup,
  title,
}) => {
  const PAGE_SIZE = 6

  const [page, setPage] = useState(0)
  const [searchValue, setSearchValue] = useState('')

  useEffect(() => {
    setPage(0)
  }, [searchValue])

  const rows = faceGroups
    .filter(
      x =>
        searchValue == '' ||
        (x.label && x.label.toLowerCase().includes(searchValue.toLowerCase()))
    )
    .map(face => (
      <FaceGroupRow
        key={face.id}
        faceGroup={face}
        faceSelected={selectedFaceGroup == face}
        setFaceSelected={() => setSelectedFaceGroup(face)}
      />
    ))

  const pageRows = rows.filter(
    (_, i) => i >= page * PAGE_SIZE && i < (page + 1) * PAGE_SIZE
  )

  return (
    <Table selectable>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell>{title}</Table.HeaderCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell>
            <Input
              value={searchValue}
              onChange={e => setSearchValue(e.target.value)}
              icon="search"
              placeholder="Search faces..."
              fluid
            />
          </Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>{pageRows}</Table.Body>
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
              totalPages={rows.length / PAGE_SIZE}
              onPageChange={(_, { activePage }) => {
                setPage(Math.ceil(activePage) - 1)
              }}
            />
          </Table.HeaderCell>
        </Table.Row>
      </Table.Footer>
    </Table>
  )
}

SelectFaceGroupTable.propTypes = {
  faceGroups: PropTypes.array,
  selectedFaceGroup: PropTypes.object,
  setSelectedFaceGroup: PropTypes.func.isRequired,
  title: PropTypes.string,
}

export default SelectFaceGroupTable
