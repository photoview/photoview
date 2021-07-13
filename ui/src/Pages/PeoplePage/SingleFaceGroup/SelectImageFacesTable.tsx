import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { ProtectedImage } from '../../../components/photoGallery/ProtectedMedia'
import Checkbox from '../../../primitives/form/Checkbox'
import { TextField } from '../../../primitives/form/Input'
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableHeaderCell,
  TableRow,
} from '../../../primitives/Table'
import { myFaces_myFaceGroups_imageFaces } from '../__generated__/myFaces'
import { singleFaceGroup_faceGroup_imageFaces } from './__generated__/singleFaceGroup'

const SelectImagePreview = styled(ProtectedImage)`
  max-width: 120px;
  max-height: 80px;
`

type ImageFaceRowProps = {
  imageFace:
    | myFaces_myFaceGroups_imageFaces
    | singleFaceGroup_faceGroup_imageFaces
  faceSelected: boolean
  setFaceSelected(): void
}

const ImageFaceRow = ({
  imageFace,
  faceSelected,
  setFaceSelected,
}: ImageFaceRowProps) => {
  return (
    <TableRow key={imageFace.id}>
      <TableCell>
        <SelectImagePreview
          src={imageFace.media.thumbnail?.url}
          onClick={setFaceSelected}
        />
      </TableCell>
      <TableCell className="min-w-64 w-full">
        <Checkbox
          label={imageFace.media.title}
          checked={faceSelected}
          onChange={setFaceSelected}
        />
      </TableCell>
    </TableRow>
  )
}

type SelectImageFacesTable = {
  imageFaces: (
    | myFaces_myFaceGroups_imageFaces
    | singleFaceGroup_faceGroup_imageFaces
  )[]
  selectedImageFaces: (
    | myFaces_myFaceGroups_imageFaces
    | singleFaceGroup_faceGroup_imageFaces
  )[]
  setSelectedImageFaces: React.Dispatch<
    React.SetStateAction<
      (myFaces_myFaceGroups_imageFaces | singleFaceGroup_faceGroup_imageFaces)[]
    >
  >
  title: string
}

const SelectImageFacesTable = ({
  imageFaces,
  selectedImageFaces,
  setSelectedImageFaces,
  title,
}: SelectImageFacesTable) => {
  const { t } = useTranslation()

  // const PAGE_SIZE = 6

  // const [page, setPage] = useState(0)
  const [searchValue, setSearchValue] = useState('')

  // useEffect(() => {
  //   setPage(0)
  // }, [searchValue])

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

  // const pageRows = rows.filter(
  //   (_, i) => i >= page * PAGE_SIZE && i < (page + 1) * PAGE_SIZE
  // )

  return (
    <>
      <Table className="w-full">
        <TableHeader>
          <TableRow>
            <TableHeaderCell colSpan={2}>{title}</TableHeaderCell>
          </TableRow>
          <TableRow>
            <TableHeaderCell colSpan={2}>
              <TextField
                value={searchValue}
                onChange={e => setSearchValue(e.target.value)}
                placeholder={t(
                  'people_page.tableselect_image_faces.search_images_placeholder',
                  'Search images...'
                )}
                fullWidth
              />
            </TableHeaderCell>
          </TableRow>
        </TableHeader>
      </Table>
      <div className="overflow-auto max-h-[500px] mt-2">
        <Table>
          <TableBody>{rows}</TableBody>
        </Table>
      </div>
    </>
  )
}

export default SelectImageFacesTable
