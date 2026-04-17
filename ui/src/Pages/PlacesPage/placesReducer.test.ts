import { placesReducer, PlacesState } from './placesReducer'
import { MediaType } from '../../__generated__/globalTypes'

const baseState: PlacesState = {
  presenting: false,
  activeIndex: 0,
  media: [],
}

describe('placesReducer', () => {
  describe('replacePresentMarker', () => {
    test('sets presentMarker when none is active', () => {
      const result = placesReducer(baseState, {
        type: 'replacePresentMarker',
        marker: { cluster: true, id: 'c1' },
      })
      expect(result.presentMarker).toEqual({ cluster: true, id: 'c1' })
      expect(result.presenting).toBe(false)
    })

    test('replaces presentMarker when a different marker is dispatched', () => {
      const state: PlacesState = {
        ...baseState,
        presentMarker: { cluster: true, id: 'c1' },
      }
      const result = placesReducer(state, {
        type: 'replacePresentMarker',
        marker: { cluster: false, id: 'm5' },
      })
      expect(result.presentMarker).toEqual({ cluster: false, id: 'm5' })
      expect(result.presenting).toBe(false)
    })

    test('sets presenting true when the same marker is dispatched again', () => {
      const state: PlacesState = {
        ...baseState,
        presentMarker: { cluster: true, id: 'c1' },
      }
      const result = placesReducer(state, {
        type: 'replacePresentMarker',
        marker: { cluster: true, id: 'c1' },
      })
      expect(result.presenting).toBe(true)
      expect(result.presentMarker).toEqual({ cluster: true, id: 'c1' })
    })

    test('does not match when cluster differs with same id', () => {
      const state: PlacesState = {
        ...baseState,
        presentMarker: { cluster: true, id: 'x' },
      }
      const result = placesReducer(state, {
        type: 'replacePresentMarker',
        marker: { cluster: false, id: 'x' },
      })
      // Should replace, not enter present mode
      expect(result.presentMarker).toEqual({ cluster: false, id: 'x' })
      expect(result.presenting).toBe(false)
    })

    test('clears marker when marker is undefined', () => {
      const state: PlacesState = {
        ...baseState,
        presentMarker: { cluster: true, id: 'c1' },
      }
      const result = placesReducer(state, {
        type: 'replacePresentMarker',
        marker: undefined,
      })
      expect(result.presentMarker).toBeUndefined()
    })
  })

  test('delegates other actions to mediaGalleryReducer', () => {
    const media = [
      {
        __typename: 'Media' as const,
        id: '1',
        highRes: null,
        thumbnail: null,
        blurhash: null,
        type: MediaType.Photo,
      },
    ]
    const result = placesReducer(baseState, {
      type: 'replaceMedia',
      media,
    })
    expect(result.media).toBe(media)
    expect(result.activeIndex).toBe(-1)
  })
})
