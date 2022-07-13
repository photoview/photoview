import React from 'react'
import { render, screen } from '@testing-library/react'
import ExifDetails from './MediaSidebarExif'
import { MediaSidebarMedia } from './MediaSidebar'
import { MediaType } from '../../../__generated__/globalTypes'

describe('ExifDetails', () => {
  test('without EXIF information', () => {
    const media: MediaSidebarMedia = {
      id: '1730',
      title: 'media_name.jpg',
      type: MediaType.Photo,
      exif: {
        id: '0',
        description: null,
        camera: null,
        maker: null,
        lens: null,
        dateShot: null,
        exposure: null,
        aperture: null,
        iso: null,
        focalLength: null,
        flash: null,
        exposureProgram: null,
        coordinates: null,
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<ExifDetails media={media} />)

    expect(screen.queryByText('Description')).not.toBeInTheDocument()
    expect(screen.queryByText('Camera')).not.toBeInTheDocument()
    expect(screen.queryByText('Maker')).not.toBeInTheDocument()
    expect(screen.queryByText('Lens')).not.toBeInTheDocument()
    expect(screen.queryByText('Program')).not.toBeInTheDocument()
    expect(screen.queryByText('Date shot')).not.toBeInTheDocument()
    expect(screen.queryByText('Exposure')).not.toBeInTheDocument()
    expect(screen.queryByText('Aperture')).not.toBeInTheDocument()
    expect(screen.queryByText('ISO')).not.toBeInTheDocument()
    expect(screen.queryByText('Focal length')).not.toBeInTheDocument()
    expect(screen.queryByText('Flash')).not.toBeInTheDocument()
    expect(screen.queryByText('Coordinates')).not.toBeInTheDocument()
  })

  test('with EXIF information', () => {
    const media: MediaSidebarMedia = {
      id: '1730',
      title: 'media_name.jpg',
      type: MediaType.Photo,
      exif: {
        id: '1666',
        description: 'Media description',
        camera: 'Canon EOS R',
        maker: 'Canon',
        lens: 'TAMRON SP 24-70mm F/2.8',
        dateShot: '2021-01-23T20:50:18Z',
        exposure: 0.016666666666666666,
        aperture: 2.8,
        iso: 100,
        focalLength: 24,
        flash: 9,
        exposureProgram: 3,
        coordinates: {
          __typename: 'Coordinates',
          latitude: 41.40338,
          longitude: 2.17403,
        },
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<ExifDetails media={media} />)

    expect(screen.getByText('Description')).toBeInTheDocument()

    expect(screen.getByText('Camera')).toBeInTheDocument()
    expect(screen.getByText('Canon EOS R')).toBeInTheDocument()

    expect(screen.getByText('Maker')).toBeInTheDocument()
    expect(screen.getByText('Canon')).toBeInTheDocument()

    expect(screen.getByText('Lens')).toBeInTheDocument()
    expect(screen.getByText('TAMRON SP 24-70mm F/2.8')).toBeInTheDocument()

    expect(screen.getByText('Program')).toBeInTheDocument()
    expect(screen.getByText('Canon EOS R')).toBeInTheDocument()

    expect(screen.getByText('Date shot')).toBeInTheDocument()

    expect(screen.getByText('Exposure')).toBeInTheDocument()
    expect(screen.getByText('1/60')).toBeInTheDocument()

    expect(screen.getByText('Program')).toBeInTheDocument()
    expect(screen.getByText('Aperture priority')).toBeInTheDocument()

    expect(screen.getByText('Aperture')).toBeInTheDocument()
    expect(screen.getByText('f/2.8')).toBeInTheDocument()

    expect(screen.getByText('ISO')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()

    expect(screen.getByText('Focal length')).toBeInTheDocument()
    expect(screen.getByText('24mm')).toBeInTheDocument()

    expect(screen.getByText('Flash')).toBeInTheDocument()
    expect(screen.getByText('On, Fired')).toBeInTheDocument()

    expect(screen.getByText('Coordinates')).toBeInTheDocument()
    expect(screen.getByText('41.40338, 2.17403')).toBeInTheDocument()
  })
})
