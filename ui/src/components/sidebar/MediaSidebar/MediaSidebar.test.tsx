import React from 'react'
import { render, screen } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import MediaSidebar, { MediaSidebarMedia } from './MediaSidebar'
import { MediaType } from '../../../__generated__/globalTypes'
import { MemoryRouter } from 'react-router'

import * as authentication from '../../../helpers/authentication'

vi.mock('../../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('MediaSidebar', () => {
  const media: MediaSidebarMedia = {
    __typename: 'Media',
    id: '6867',
    title: '122A6069.jpg',
    type: MediaType.Photo,
    thumbnail: {
      __typename: 'MediaURL',
      url: 'http://localhost:4001/photo/thumbnail.jpg',
      width: 1024,
      height: 839,
    },
    highRes: {
      __typename: 'MediaURL',
      url: 'http://localhost:4001/photo/highres.jpg',
      width: 5322,
      height: 4362,
    },
    videoWeb: null,
    album: {
      __typename: 'Album',
      id: '2294',
      title: 'album_name',
    },
  }

  test('render sample image, unauthorized', () => {
    authToken.mockImplementation(() => null)

    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <MemoryRouter>
          <MediaSidebar media={media} />
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('122A6069.jpg')).toBeInTheDocument()
    expect(screen.getByRole('img')).toHaveAttribute(
      'src',
      'http://localhost:4001/photo/highres.jpg'
    )

    expect(
      screen.queryByText('Set as album cover photo')
    ).not.toBeInTheDocument()
    expect(screen.queryByText('Sharing options')).not.toBeInTheDocument()
  })

  test('render sample image, authorized', () => {
    authToken.mockImplementation(() => 'token-here')

    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <MemoryRouter>
          <MediaSidebar media={media} />
        </MemoryRouter>
      </MockedProvider>
    )

    expect(screen.getByText('122A6069.jpg')).toBeInTheDocument()
    expect(screen.getByRole('img')).toHaveAttribute(
      'src',
      'http://localhost:4001/photo/highres.jpg'
    )

    expect(screen.getByText('Set as album cover photo')).toBeInTheDocument()
    expect(screen.getByText('Album path')).toBeInTheDocument()

    screen.debug()
  })
})
