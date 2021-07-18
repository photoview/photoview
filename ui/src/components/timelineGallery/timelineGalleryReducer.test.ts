import { myTimeline_myTimeline } from './__generated__/myTimeline'
import { MediaType } from '../../__generated__/globalTypes'
import {
  timelineGalleryReducer,
  TimelineGalleryState,
  TimelineMediaIndex,
} from './timelineGalleryReducer'

describe('timeline gallery reducer', () => {
  const defaultEmptyState: TimelineGalleryState = {
    presenting: false,
    activeIndex: {
      album: -1,
      date: -1,
      media: -1,
    },
    timelineGroups: [],
  }

  const timelineData: myTimeline_myTimeline[] = [
    {
      album: {
        id: '5',
        title: 'first album',
        __typename: 'Album',
      },
      media: [
        {
          id: '165',
          title: '3666760020.jpg',
          type: MediaType.Photo,
          thumbnail: {
            url: 'http://localhost:4001/photo/thumbnail_3666760020_jpg_x76GG5pS.jpg',
            width: 768,
            height: 1024,
            __typename: 'MediaURL',
          },
          highRes: {
            url: 'http://localhost:4001/photo/3666760020_wijGDNZ2.jpg',
            width: 3024,
            height: 4032,
            __typename: 'MediaURL',
          },
          videoWeb: null,
          favorite: false,
          __typename: 'Media',
        },
        {
          id: '184',
          title: '7414455077.jpg',
          type: MediaType.Photo,
          thumbnail: {
            url: 'http://localhost:4001/photo/thumbnail_7414455077_jpg_9JYHHYh6.jpg',
            width: 768,
            height: 1024,
            __typename: 'MediaURL',
          },
          highRes: {
            url: 'http://localhost:4001/photo/7414455077_0ejDBiKr.jpg',
            width: 3024,
            height: 4032,
            __typename: 'MediaURL',
          },
          videoWeb: null,
          favorite: false,
          __typename: 'Media',
        },
      ],
      mediaTotal: 5,
      date: '2019-09-21T00:00:00Z',
      __typename: 'TimelineGroup',
    },
    {
      album: {
        id: '5',
        title: 'another album',
        __typename: 'Album',
      },
      media: [
        {
          id: '165',
          title: '3666760020.jpg',
          type: MediaType.Photo,
          thumbnail: {
            url: 'http://localhost:4001/photo/thumbnail_3666760020_jpg_x76GG5pS.jpg',
            width: 768,
            height: 1024,
            __typename: 'MediaURL',
          },
          highRes: {
            url: 'http://localhost:4001/photo/3666760020_wijGDNZ2.jpg',
            width: 3024,
            height: 4032,
            __typename: 'MediaURL',
          },
          videoWeb: null,
          favorite: false,
          __typename: 'Media',
        },
      ],
      mediaTotal: 7,
      date: '2019-09-21T00:00:00Z',
      __typename: 'TimelineGroup',
    },
    {
      __typename: 'TimelineGroup',
      album: {
        __typename: 'Album',
        id: '5',
        title: 'album on another day',
      },
      date: '2019-09-13T00:00:00Z',
      mediaTotal: 1,
      media: [
        {
          __typename: 'Media',
          favorite: false,
          videoWeb: null,
          thumbnail: {
            url: 'http://localhost:4001/photo/thumbnail_3666760020_jpg_x76GG5pS.jpg',
            width: 768,
            height: 1024,
            __typename: 'MediaURL',
          },
          highRes: {
            url: 'http://localhost:4001/photo/3666760020_wijGDNZ2.jpg',
            width: 3024,
            height: 4032,
            __typename: 'MediaURL',
          },
          id: '321',
          title: 'asdfimg.jpg',
          type: MediaType.Photo,
        },
      ],
    },
  ]

  const defaultState = timelineGalleryReducer(defaultEmptyState, {
    type: 'replaceTimelineGroups',
    timeline: timelineData,
  })

  test('replace timeline groups', () => {
    expect(defaultState).toMatchObject({
      presenting: false,
      activeIndex: {
        album: -1,
        date: -1,
        media: -1,
      },
      timelineGroups: [
        {
          date: '2019-09-21T00:00:00Z',
          groups: [
            {
              album: {
                id: '5',
                title: 'first album',
              },
              date: '2019-09-21T00:00:00Z',
              media: [
                {
                  favorite: false,
                  highRes: {},
                  id: '165',
                  thumbnail: {},
                  title: '3666760020.jpg',
                  type: 'Photo',
                },
                {
                  highRes: {},
                  id: '184',
                  thumbnail: {},
                  title: '7414455077.jpg',
                  type: 'Photo',
                },
              ],
              mediaTotal: 5,
            },
            {
              album: {
                id: '5',
                title: 'another album',
              },
              date: '2019-09-21T00:00:00Z',
              media: [
                {
                  id: '165',
                },
              ],
              mediaTotal: 7,
            },
          ],
        },
        {
          date: '2019-09-13T00:00:00Z',
          groups: [
            {
              album: {
                id: '5',
                title: 'album on another day',
              },
              date: '2019-09-13T00:00:00Z',
              media: [
                {
                  favorite: false,
                  highRes: {},
                  id: '321',
                  thumbnail: {},
                  title: 'asdfimg.jpg',
                  type: 'Photo',
                },
              ],
              mediaTotal: 1,
            },
          ],
        },
      ],
    })
  })

  test('select image', () => {
    expect(
      timelineGalleryReducer(defaultState, {
        type: 'selectImage',
        index: {
          album: 0,
          date: 0,
          media: 0,
        },
      })
    ).toEqual({
      ...defaultState,
      activeIndex: {
        album: 0,
        date: 0,
        media: 0,
      },
    })
  })

  describe('next image', () => {
    const testIndexes: {
      name: string
      in: TimelineMediaIndex
      out: TimelineMediaIndex
    }[] = [
      {
        name: 'no selection',
        in: {
          date: -1,
          album: -1,
          media: -1,
        },
        out: {
          date: -1,
          album: -1,
          media: -1,
        },
      },
      {
        name: 'first selected',
        in: {
          date: 0,
          album: 0,
          media: 0,
        },
        out: {
          date: 0,
          album: 0,
          media: 1,
        },
      },
      {
        name: 'next album',
        in: {
          date: 0,
          album: 0,
          media: 1,
        },
        out: {
          date: 0,
          album: 1,
          media: 0,
        },
      },
      {
        name: 'next date',
        in: {
          date: 0,
          album: 1,
          media: 1,
        },
        out: {
          date: 1,
          album: 0,
          media: 0,
        },
      },
      {
        name: 'reached end',
        in: {
          date: 1,
          album: 0,
          media: 0,
        },
        out: {
          date: 1,
          album: 0,
          media: 0,
        },
      },
    ]

    testIndexes.forEach(t => {
      test(t.name, () => {
        expect(
          timelineGalleryReducer(
            {
              ...defaultState,
              activeIndex: t.in,
            },
            { type: 'nextImage' }
          )
        ).toEqual({
          ...defaultState,
          activeIndex: t.out,
        })
      })
    })
  })

  describe('previous image', () => {
    const testIndexes: {
      name: string
      in: TimelineMediaIndex
      out: TimelineMediaIndex
    }[] = [
      {
        name: 'no selection',
        in: {
          album: -1,
          date: -1,
          media: -1,
        },
        out: {
          album: -1,
          date: -1,
          media: -1,
        },
      },
      {
        name: 'first selected',
        in: {
          album: 0,
          date: 0,
          media: 0,
        },
        out: {
          album: 0,
          date: 0,
          media: 0,
        },
      },
      {
        name: 'previous album',
        in: {
          date: 0,
          album: 1,
          media: 0,
        },
        out: {
          date: 0,
          album: 0,
          media: 1,
        },
      },
      {
        name: 'previous date',
        in: {
          date: 1,
          album: 0,
          media: 0,
        },
        out: {
          date: 0,
          album: 1,
          media: 0,
        },
      },
      {
        name: 'previous media',
        in: {
          date: 0,
          album: 0,
          media: 1,
        },
        out: {
          date: 0,
          album: 0,
          media: 0,
        },
      },
    ]

    testIndexes.forEach(t => {
      test(t.name, () => {
        expect(
          timelineGalleryReducer(
            {
              ...defaultState,
              activeIndex: t.in,
            },
            { type: 'previousImage' }
          )
        ).toEqual({
          ...defaultState,
          activeIndex: t.out,
        })
      })
    })
  })

  test('present mode', () => {
    expect(
      timelineGalleryReducer(defaultState, {
        type: 'openPresentMode',
        activeIndex: {
          date: 1,
          album: 0,
          media: 0,
        },
      })
    ).toEqual({
      ...defaultState,
      presenting: true,
      activeIndex: {
        date: 1,
        album: 0,
        media: 0,
      },
    })

    expect(
      timelineGalleryReducer(defaultState, {
        type: 'closePresentMode',
      })
    ).toEqual({
      ...defaultState,
      presenting: false,
    })
  })
})
