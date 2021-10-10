import React from 'react'
import { render, screen } from '@testing-library/react'
import { MetadataInfo } from './MediaSidebar'

describe('MetadataInfo', () => {
  test('without EXIF information', async () => {
    const media = {
      id: '1730',
      title: 'media_name.jpg',
      type: 'Photo',
      exif: {
        id: '0',
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
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<MetadataInfo media={media} />)

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
  })

  test('with EXIF information', async () => {
    const media = {
      id: '1730',
      title: 'media_name.jpg',
      type: 'Photo',
      exif: {
        id: '1666',
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
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<MetadataInfo media={media} />)

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
  })
})
