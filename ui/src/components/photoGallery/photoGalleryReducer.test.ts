import { photoGalleryReducer, PhotoGalleryState } from './photoGalleryReducer'
import { MediaType } from '../../../__generated__/globalTypes'

describe('photo gallery reducer', () => {
  const defaultState: PhotoGalleryState = {
    presenting: false,
    activeIndex: 0,
    media: [
      {
        __typename: 'Media',
        id: '1',
        highRes: null,
        thumbnail: null,
        type: MediaType.Photo,
      },
      {
        __typename: 'Media',
        id: '2',
        highRes: null,
        thumbnail: null,
        type: MediaType.Photo,
      },
      {
        __typename: 'Media',
        id: '3',
        highRes: null,
        thumbnail: null,
        type: MediaType.Photo,
      },
    ],
  }

  test('next image', () => {
    expect(photoGalleryReducer(defaultState, { type: 'nextImage' })).toEqual({
      ...defaultState,
      activeIndex: 1,
    })

    expect(
      photoGalleryReducer(
        { ...defaultState, activeIndex: 2 },
        { type: 'nextImage' }
      )
    ).toEqual({
      ...defaultState,
      activeIndex: 0,
    })
  })

  test('previous image', () => {
    expect(
      photoGalleryReducer(defaultState, { type: 'previousImage' })
    ).toEqual({
      ...defaultState,
      activeIndex: 2,
    })

    expect(
      photoGalleryReducer(
        { ...defaultState, activeIndex: 2 },
        { type: 'previousImage' }
      )
    ).toEqual({
      ...defaultState,
      activeIndex: 1,
    })
  })

  test('select image', () => {
    expect(
      photoGalleryReducer(defaultState, { type: 'selectImage', index: 1 })
    ).toEqual({
      ...defaultState,
      activeIndex: 1,
    })

    expect(
      photoGalleryReducer(defaultState, { type: 'selectImage', index: 100 })
    ).toEqual({
      ...defaultState,
      activeIndex: 2,
    })

    expect(
      photoGalleryReducer(defaultState, { type: 'selectImage', index: -5 })
    ).toEqual({
      ...defaultState,
      activeIndex: 0,
    })
  })
})
