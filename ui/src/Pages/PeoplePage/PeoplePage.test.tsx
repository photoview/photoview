import '@testing-library/jest-dom'

import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import PeoplePage, { MY_FACES_QUERY } from './PeoplePage'
import { MockedProvider } from '@apollo/client/testing'
import { MemoryRouter } from 'react-router'

require('../../localization').setupLocalization()

jest.mock('../../hooks/useScrollPagination', () =>
  jest.fn(() => ({
    finished: true,
    containerElem: jest.fn(),
  }))
)

const graphqlMocks = [
  {
    request: {
      query: MY_FACES_QUERY,
      variables: {
        limit: 50,
        offset: 0,
      },
    },
    result: {
      data: {
        myFaceGroups: [
          {
            __typename: 'FaceGroup',
            id: '3',
            label: 'Person A',
            imageFaceCount: 2,
            imageFaces: [
              {
                __typename: 'ImageFace',
                id: '3',
                rectangle: {
                  __typename: 'FaceRectangle',
                  minX: 0.2705079913139343,
                  maxX: 0.3408200144767761,
                  minY: 0.7691109776496887,
                  maxY: 0.881434977054596,
                },
                media: {
                  __typename: 'Media',
                  id: '63',
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/thumbnail_L%C3%B8berute_jpg_p9x8dLWr.jpg',
                    width: 1024,
                    height: 641,
                  },
                },
              },
            ],
          },
          {
            __typename: 'FaceGroup',
            id: '1',
            label: 'Person B',
            imageFaceCount: 1,
            imageFaces: [],
          },
        ],
      },
    },
  },
]

test('people page', async () => {
  const matchMock = {
    params: {
      person: undefined,
    },
  }

  render(
    <MemoryRouter initialEntries={['/people']}>
      <MockedProvider mocks={graphqlMocks} addTypename={false}>
        <PeoplePage match={matchMock} />
      </MockedProvider>
    </MemoryRouter>
  )

  expect(screen.getByTestId('Layout')).toBeInTheDocument()
  expect(screen.getByText('Recognize unlabeled faces')).toBeInTheDocument()

  await waitFor(() => {
    expect(screen.getByText('Person A')).toBeInTheDocument()
    expect(screen.getByText('Person B')).toBeInTheDocument()
  })

  expect(
    screen.getAllByRole('link').some(x => x.getAttribute('href') == '/people/1')
  ).toBeTruthy()

  expect(
    screen.getAllByRole('link').some(x => x.getAttribute('href') == '/people/3')
  ).toBeTruthy()
})
