import React from 'react'
import { render, screen } from '@testing-library/react'
import FaceCircleImage from './FaceCircleImage'
import { myFaces_myFaceGroups_imageFaces } from './__generated__/myFaces'

test('face circle image', () => {
  const imageFace: myFaces_myFaceGroups_imageFaces = {
    __typename: 'ImageFace',
    id: '3',
    media: {
      id: '1',
      __typename: 'Media',
      title: 'my_image.jpg',
      thumbnail: {
        __typename: 'MediaURL',
        url: 'http://localhost:4001/photo/thumbnail_my_image_jpg_p9x8dLWr.jpg',
        width: 1024,
        height: 641,
      },
    },
    rectangle: {
      __typename: 'FaceRectangle',
      minX: 0.27,
      maxX: 0.34,
      minY: 0.76,
      maxY: 0.88,
    },
  }

  render(<FaceCircleImage imageFace={imageFace} selectable={true} />)

  expect(screen.getByRole('img')).toBeInTheDocument()
  expect(screen.getByRole('img')).toHaveAttribute(
    'src',
    imageFace.media.thumbnail!.url
  )
})
