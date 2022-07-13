import { mediaGalleryReducer, MediaGalleryState } from './mediaGalleryReducer'
import { MediaType } from '../../__generated__/globalTypes'

describe('photo gallery reducer', () => {
  const defaultState: MediaGalleryState = {
    presenting: false,
    activeIndex: 0,
    media: [
      {
        __typename: 'Media',
        id: '1',
        highRes: null,
        thumbnail: null,
        blurhash: null,
        type: MediaType.Photo,
      },
      {
        __typename: 'Media',
        id: '2',
        highRes: null,
        thumbnail: null,
        blurhash: null,
        type: MediaType.Photo,
      },
      {
        __typename: 'Media',
        id: '3',
        highRes: null,
        thumbnail: null,
        blurhash: null,
        type: MediaType.Photo,
      },
    ],
  }

  test('next image', () => {
    expect(mediaGalleryReducer(defaultState, { type: 'nextImage' })).toEqual({
      ...defaultState,
      activeIndex: 1,
    })

    expect(
      mediaGalleryReducer(
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
      mediaGalleryReducer(defaultState, { type: 'previousImage' })
    ).toEqual({
      ...defaultState,
      activeIndex: 2,
    })

    expect(
      mediaGalleryReducer(
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
      mediaGalleryReducer(defaultState, { type: 'selectImage', index: 1 })
    ).toEqual({
      ...defaultState,
      activeIndex: 1,
    })

    expect(
      mediaGalleryReducer(defaultState, { type: 'selectImage', index: 100 })
    ).toEqual({
      ...defaultState,
      activeIndex: 2,
    })

    expect(
      mediaGalleryReducer(defaultState, { type: 'selectImage', index: -5 })
    ).toEqual({
      ...defaultState,
      activeIndex: 0,
    })
  })

  test('present mode', () => {
    const openState = mediaGalleryReducer(defaultState, {
      type: 'openPresentMode',
      activeIndex: 10,
    })
    expect(openState).toEqual({
      ...defaultState,
      presenting: true,
      activeIndex: 10,
    })

    expect(
      mediaGalleryReducer(openState, { type: 'closePresentMode' })
    ).toEqual({
      ...openState,
      presenting: false,
    })
  })
})
