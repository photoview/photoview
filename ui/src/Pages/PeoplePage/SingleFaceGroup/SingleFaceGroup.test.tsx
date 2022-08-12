import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import SingleFaceGroup, { SINGLE_FACE_GROUP } from './SingleFaceGroup'
import { MemoryRouter } from 'react-router-dom'

vi.mock('../../../hooks/useScrollPagination')

test('single face group', async () => {
  const graphqlMocks = [
    {
      request: {
        query: SINGLE_FACE_GROUP,
        variables: { limit: 200, offset: 0, id: '123' },
      },
      result: {
        data: {
          faceGroup: {
            __typename: 'FaceGroup',
            id: '2',
            label: 'Face Group Name',
            imageFaces: [
              {
                __typename: 'ImageFace',
                id: '1',
                rectangle: {
                  __typename: 'FaceRectangle',
                  minX: 0.4912109971046448,
                  maxX: 0.5927730202674866,
                  minY: 0.2998049855232239,
                  maxY: 0.4013670086860657,
                },
                media: {
                  __typename: 'Media',
                  id: '10',
                  type: 'Photo',
                  title: '122A2785-2.jpg',
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/thumbnail_122A2785-2_jpg_lFmZcaN5.jpg',
                    width: 1024,
                    height: 1024,
                  },
                  highRes: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/122A2785-2_e4nCeMHU.jpg',
                  },
                  favorite: false,
                },
              },
              {
                __typename: 'ImageFace',
                id: '2',
                rectangle: {
                  __typename: 'FaceRectangle',
                  minX: 0.265625,
                  maxX: 0.3876950144767761,
                  minY: 0.1917019933462143,
                  maxY: 0.3705289959907532,
                },
                media: {
                  __typename: 'Media',
                  id: '52',
                  type: 'Photo',
                  title: 'image.png',
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/thumbnail_image_png_OwTDG5fM.jpg',
                    width: 1024,
                    height: 699,
                  },
                  highRes: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/image_A2YB0x3z.png',
                  },
                  favorite: false,
                },
              },
            ],
          },
        },
      },
    },
  ]

  render(
    <MemoryRouter initialEntries={['/person/123']}>
      <MockedProvider mocks={graphqlMocks}>
        <SingleFaceGroup faceGroupID="123" />
      </MockedProvider>
    </MemoryRouter>
  )

  await waitFor(() => {
    expect(screen.getAllByRole('img')).toHaveLength(2)
  })
})
