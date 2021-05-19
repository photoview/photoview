import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Input, Pagination, Table } from 'semantic-ui-react'
import styled from 'styled-components'
import FaceCircleImage from '../FaceCircleImage'
import { myFaces_myFaceGroups } from '../__generated__/myFaces'
import { singleFaceGroup_faceGroup } from './__generated__/singleFaceGroup'

const FaceCircleWrapper = styled.div<{ $selected: boolean }>`
  display: inline-block;
  border-radius: 50%;
  border: 2px solid
    ${({ $selected }) => ($selected ? `#2185c9` : 'rgba(255,255,255,0)')};
`

const FlexCell = styled(Table.Cell)`
  display: flex;
  align-items: center;
`

export const RowLabel = styled.span<{ $selected: boolean }>`
  ${({ $selected }) => $selected && `font-weight: bold;`}
  margin-left: 12px;
`

type FaceGroupRowProps = {
  faceGroup: myFaces_myFaceGroups
  faceSelected: boolean
  setFaceSelected(): void
}

const FaceGroupRow = ({
  faceGroup,
  faceSelected,
  setFaceSelected,
}: FaceGroupRowProps) => {
  return (
    <Table.Row key={faceGroup.id} onClick={setFaceSelected}>
      <FlexCell>
        <FaceCircleWrapper $selected={faceSelected}>
          <FaceCircleImage
            imageFace={faceGroup.imageFaces[0]}
            size="50px"
            selectable={false}
          />
        </FaceCircleWrapper>
        <RowLabel $selected={faceSelected}>{faceGroup.label}</RowLabel>
      </FlexCell>
    </Table.Row>
  )
}

type SelectFaceGroupTableProps = {
  faceGroups: myFaces_myFaceGroups[]
  selectedFaceGroup: singleFaceGroup_faceGroup | myFaces_myFaceGroups | null
  setSelectedFaceGroup: React.Dispatch<
    React.SetStateAction<
      singleFaceGroup_faceGroup | myFaces_myFaceGroups | null
    >
  >
  title: string
}

const SelectFaceGroupTable = ({
  faceGroups,
  selectedFaceGroup,
  setSelectedFaceGroup,
  title,
}: SelectFaceGroupTableProps) => {
  const { t } = useTranslation()

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
        faceSelected={selectedFaceGroup?.id == face.id}
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
              placeholder={t(
                'people_page.table.select_face_group.search_faces_placeholder',
                'Search faces...'
              )}
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
              totalPages={Math.ceil(rows.length / PAGE_SIZE)}
              onPageChange={(_, { activePage }) => {
                if (activePage) {
                  setPage(Math.ceil(activePage as number) - 1)
                } else {
                  setPage(0)
                }
              }}
            />
          </Table.HeaderCell>
        </Table.Row>
      </Table.Footer>
    </Table>
  )
}

export default SelectFaceGroupTable
