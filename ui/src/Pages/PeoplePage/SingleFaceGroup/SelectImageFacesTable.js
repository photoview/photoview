import PropTypes from 'prop-types'
import React, { useEffect, useState } from 'react'
import { Checkbox, Input, Pagination, Table } from 'semantic-ui-react'
import styled from 'styled-components'
import { ProtectedImage } from '../../../components/photoGallery/ProtectedMedia'
import { RowLabel } from './SelectFaceGroupTable'

const SelectImagePreview = styled(ProtectedImage)`
  max-width: 120px;
  max-height: 80px;
`

const ImageFaceRow = ({ imageFace, faceSelected, setFaceSelected }) => {
  return (
    <Table.Row key={imageFace.id}>
      <Table.Cell>
        <Checkbox checked={faceSelected} onChange={setFaceSelected} />
      </Table.Cell>
      <Table.Cell>
        <SelectImagePreview
          src={imageFace.media.thumbnail.url}
          onClick={setFaceSelected}
        />
      </Table.Cell>
      <Table.Cell width={16}>
        <RowLabel $selected={faceSelected} onClick={setFaceSelected}>
          {imageFace.media.title}
        </RowLabel>
      </Table.Cell>
    </Table.Row>
  )
}

ImageFaceRow.propTypes = {
  imageFace: PropTypes.object.isRequired,
  faceSelected: PropTypes.bool.isRequired,
  setFaceSelected: PropTypes.func.isRequired,
}

const SelectImageFacesTable = ({
  imageFaces,
  selectedImageFaces,
  setSelectedImageFaces,
  title,
}) => {
  const PAGE_SIZE = 6

  const [page, setPage] = useState(0)
  const [searchValue, setSearchValue] = useState('')

  useEffect(() => {
    setPage(0)
  }, [searchValue])

  const rows = imageFaces
    .filter(
      face =>
        searchValue == '' ||
        face.media.title.toLowerCase().includes(searchValue.toLowerCase())
    )
    .map(face => (
      <ImageFaceRow
        key={face.id}
        imageFace={face}
        faceSelected={selectedImageFaces.includes(face)}
        setFaceSelected={() =>
          setSelectedImageFaces(faces => {
            if (faces.includes(face)) {
              return faces.filter(x => x != face)
            } else {
              return [...faces, face]
            }
          })
        }
      />
    ))

  const pageRows = rows.filter(
    (_, i) => i >= page * PAGE_SIZE && i < (page + 1) * PAGE_SIZE
  )

  return (
    <Table selectable>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell colSpan={3}>{title}</Table.HeaderCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell colSpan={3}>
            <Input
              value={searchValue}
              onChange={e => setSearchValue(e.target.value)}
              icon="search"
              placeholder="Search images..."
              fluid
            />
          </Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>{pageRows}</Table.Body>
      <Table.Footer>
        <Table.Row>
          <Table.HeaderCell colSpan={3}>
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

SelectImageFacesTable.propTypes = {
  imageFaces: PropTypes.array,
  selectedImageFaces: PropTypes.array,
  setSelectedImageFaces: PropTypes.func.isRequired,
  title: PropTypes.string,
}

export default SelectImageFacesTable
