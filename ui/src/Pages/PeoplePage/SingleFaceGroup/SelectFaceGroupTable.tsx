import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { TextField } from '../../../primitives/form/Input'
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableHeaderCell,
  TableRow,
} from '../../../primitives/Table'
import FaceCircleImage from '../FaceCircleImage'
import { myFaces_myFaceGroups } from '../__generated__/myFaces'
import { singleFaceGroup_faceGroup } from './__generated__/singleFaceGroup'

const FaceCircleWrapper = styled.div<{ $selected: boolean }>`
  display: inline-block;
  border-radius: 50%;
  border: 2px solid
    ${({ $selected }) => ($selected ? `#2185c9` : 'rgba(255,255,255,0)')};
`

const FlexCell = styled(TableCell)`
  display: flex;
  align-items: center;
`

export const RowLabel = styled.span<{ $selected: boolean }>`
  ${({ $selected }) => $selected && `font-weight: bold;`}
  margin-left: 12px;
  width: 100%;
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
    <TableRow key={faceGroup.id} onClick={setFaceSelected}>
      <FlexCell>
        <FaceCircleWrapper $selected={faceSelected}>
          <FaceCircleImage
            imageFace={faceGroup.imageFaces[0]}
            size="50px"
            selectable={false}
          />
        </FaceCircleWrapper>
        <span
          className={`ml-3 ${faceSelected ? 'font-semibold' : ''} ${
            !faceGroup.label ? 'text-gray-500 italic' : ''
          }`}
        >
          {faceGroup.label ?? 'Unlabeled'}
        </span>
      </FlexCell>
    </TableRow>
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

  const [searchValue, setSearchValue] = useState('')

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

  return (
    <>
      <Table className="w-full">
        <TableHeader>
          <TableRow>
            <TableHeaderCell>{title}</TableHeaderCell>
          </TableRow>
          <TableRow>
            <TableHeaderCell>
              <TextField
                fullWidth
                value={searchValue}
                onChange={e => setSearchValue(e.target.value)}
                placeholder={t(
                  'people_page.tableselect_face_group.search_faces_placeholder',
                  'Search faces...'
                )}
              />
            </TableHeaderCell>
          </TableRow>
        </TableHeader>
      </Table>
      <div className="overflow-auto max-h-[500px] mt-2">
        <Table className="w-full">
          <TableBody>{rows}</TableBody>
        </Table>
      </div>
    </>
  )
}

export default SelectFaceGroupTable
