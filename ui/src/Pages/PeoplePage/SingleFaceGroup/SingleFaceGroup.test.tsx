import { screen, waitFor } from '@testing-library/react'
import SingleFaceGroup, { SINGLE_FACE_GROUP } from './SingleFaceGroup'
import { renderWithProviders } from '../../../helpers/testUtils'
import { MY_FACES_QUERY } from '../PeoplePage'

vi.mock('react-blurhash', () => ({
  Blurhash: () =>
    <div data-testid="mock-blurhash">Blurhash</div>,
  BlurhashCanvas: () =>
    <div data-testid="mock-blurhash-canvas">BlurhashCanvas</div>
}))
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
                  blurhash: 'LKO2?U%2Tw=w]~RBVZRi};RPxuwH',
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: '/photo/thumbnail_122A2785-2_jpg_lFmZcaN5.jpg',
                    width: 1024,
                    height: 1024,
                  },
                  highRes: {
                    __typename: 'MediaURL',
                    url: '/photo/122A2785-2_e4nCeMHU.jpg',
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
                  blurhash: 'LKO2?U%2Tw=w]~RBVZRi};RPxuwH',
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: '/photo/thumbnail_image_png_OwTDG5fM.jpg',
                    width: 1024,
                    height: 699,
                  },
                  highRes: {
                    __typename: 'MediaURL',
                    url: '/photo/image_A2YB0x3z.png',
                  },
                  favorite: false,
                },
              },
            ],
          },
        },
      },
    },
    {
      request: {
        query: MY_FACES_QUERY,
        variables: {},
      },
      result: {
        data: {
          myFaceGroups: []
        }
      }
    }
  ]

  renderWithProviders(<SingleFaceGroup faceGroupID="123" />, {
    mocks: graphqlMocks,
    initialEntries: ['/person/123']
  })

  await waitFor(() => {
    const blurhashElements = screen.queryAllByTestId('mock-blurhash');
    const canvasElements = screen.queryAllByTestId('mock-blurhash-canvas');
    expect(blurhashElements.length + canvasElements.length).toBeGreaterThan(0);
  }, { timeout: 2000 });
})

test('handles GraphQL error', async () => {
  const graphqlMocks = [
    {
      request: {
        query: SINGLE_FACE_GROUP,
        variables: { limit: 200, offset: 0, id: '123' },
      },
      error: new Error('An error occurred'),
    },
    {
      request: {
        query: MY_FACES_QUERY,
        variables: {},
      },
      result: {
        data: {
          myFaceGroups: []
        }
      }
    }
  ]

  renderWithProviders(<SingleFaceGroup faceGroupID="123" />, {
    mocks: graphqlMocks,
    initialEntries: ['/person/123']
  })

  // Check that the error message is displayed
  await waitFor(() => {
    expect(screen.getByText('An error occurred')).toBeInTheDocument();
  }, { timeout: 2000 });
})

test('handles face group not found', async () => {
  const graphqlMocks = [
    {
      request: {
        query: SINGLE_FACE_GROUP,
        variables: { limit: 200, offset: 0, id: '123' },
      },
      result: {
        data: {
          faceGroup: null,
        },
      },
    },
    {
      request: {
        query: MY_FACES_QUERY,
        variables: {},
      },
      result: {
        data: {
          myFaceGroups: []
        }
      }
    }
  ]

  renderWithProviders(<SingleFaceGroup faceGroupID="123" />, {
    mocks: graphqlMocks,
    initialEntries: ['/person/123']
  })

  await waitFor(() => {
    expect(screen.getByText('Face group not found')).toBeInTheDocument();
  }, { timeout: 2000 });
})
