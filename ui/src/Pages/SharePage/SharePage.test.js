import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import { MockedProvider } from 'react-apollo/testing'
import { create } from 'react-test-renderer'

import SharePage, { SHARE_TOKEN_QUERY } from './SharePage'

test('two plus two is four', () => {
  const token = 'token_here'

  const matchMock = {
    url: `/share`,
    params: {},
    path: `/share`,
    isExact: false,
  }

  const historyMock = [{ pathname: `/share/${token}` }]

  const graphqlMocks = [
    {
      request: {
        query: SHARE_TOKEN_QUERY,
        variables: {
          token,
          password: null,
        },
      },
      response: {
        data: {
          shareToken: {
            token: token,
            album: null,
            media: {
              id: 1,
              title: 'shared_image.jpg',
              type: 'photo',
            },
          },
        },
      },
    },
  ]

  const mediaSharePage = create(
    <MockedProvider mocks={graphqlMocks} addTypename={false}>
      <MemoryRouter initialEntries={historyMock}>
        <SharePage match={matchMock} />
      </MemoryRouter>
    </MockedProvider>
  )

  console.log(mediaSharePage.toJSON())
})
