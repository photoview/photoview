import { render, screen } from '@testing-library/react'

import React from 'react'
import { MediaType } from '../../../__generated__/globalTypes'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'
import PresentMedia from './PresentMedia'

test('render present image', () => {
  const media: MediaGalleryFields = {
    __typename: 'Media',
    id: '123',
    type: MediaType.Photo,
    highRes: null,
    blurhash: null,
    videoWeb: null,
    favorite: false,
    thumbnail: {
      __typename: 'MediaURL',
      url: '/sample_image.jpg',
      width: 300,
      height: 200,
    },
  }

  render(<PresentMedia media={media} />)

  expect(screen.getByTestId('present-img-thumbnail')).toHaveAttribute(
    'src',
    'http://localhost:3000/sample_image.jpg'
  )
  expect(screen.getByTestId('present-img-highres')).toHaveStyle({
    display: 'none',
  })
})

test('render present video', () => {
  const media: MediaGalleryFields = {
    __typename: 'Media',
    id: '123',
    type: MediaType.Video,
    highRes: null,
    blurhash: null,
    favorite: false,
    videoWeb: {
      __typename: 'MediaURL',
      url: '/sample_video.mp4',
    },
    thumbnail: {
      __typename: 'MediaURL',
      url: '/sample_video_thumb.jpg',
      width: 300,
      height: 200,
    },
  }

  render(<PresentMedia media={media} />)

  expect(screen.getByTestId('present-video')).toHaveAttribute(
    'poster',
    'http://localhost:3000/sample_video_thumb.jpg'
  )

  expect(
    screen.getByTestId('present-video').querySelector('source')
  ).toHaveAttribute('src', 'http://localhost:3000/sample_video.mp4')
})
