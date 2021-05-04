import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'

import React from 'react'
import { MediaType } from '../../../../__generated__/globalTypes'
import PresentMedia, { PresentMediaProps_Media } from './PresentMedia'

test('render present image', () => {
  const media: PresentMediaProps_Media = {
    __typename: 'Media',
    id: '123',
    type: MediaType.Photo,
    highRes: null,
    thumbnail: {
      __typename: 'MediaURL',
      url: '/sample_image.jpg',
    },
  }

  render(<PresentMedia media={media} />)

  expect(screen.getByTestId('present-img-thumbnail')).toHaveAttribute(
    'src',
    'http://localhost/sample_image.jpg'
  )
  expect(screen.getByTestId('present-img-highres')).toHaveStyle({
    display: 'none',
  })
})

test('render present video', () => {
  const media: PresentMediaProps_Media = {
    __typename: 'Media',
    id: '123',
    type: MediaType.Video,
    highRes: null,
    videoWeb: {
      __typename: 'MediaURL',
      url: '/sample_video.mp4',
    },
    thumbnail: {
      __typename: 'MediaURL',
      url: '/sample_video_thumb.jpg',
    },
  }

  render(<PresentMedia media={media} />)

  expect(screen.getByTestId('present-video')).toHaveAttribute(
    'poster',
    'http://localhost/sample_video_thumb.jpg'
  )

  expect(
    screen.getByTestId('present-video').querySelector('source')
  ).toHaveAttribute('src', 'http://localhost/sample_video.mp4')
})
