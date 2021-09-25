import { MediaType } from '../../__generated__/globalTypes'
import {
  timelineGalleryReducer,
  TimelineGalleryState,
  TimelineMediaIndex,
} from './timelineGalleryReducer'
import { timelineData } from './timelineTestData'

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
          date: '2020-12-13T00:00:00Z',
          albums: [
            {
              id: '522',
              media: [
                {
                  __typename: 'Media',
                  album: {
                    __typename: 'Album',
                    id: '522',
                    title: 'random',
                  },
                  date: '2020-12-13T18:03:40Z',
                  favorite: false,
                  highRes: {
                    __typename: 'MediaURL',
                    height: 4480,
                    url: 'http://localhost:4001/photo/122A2876_5cSPMiKL.jpg',
                    width: 6720,
                  },
                  id: '1058',
                  thumbnail: {
                    __typename: 'MediaURL',
                    height: 682,
                    url: 'http://localhost:4001/photo/thumbnail_122A2876_jpg_Kp1U80vD.jpg',
                    width: 1024,
                  },
                  title: '122A2876.jpg',
                  type: 'Photo',
                  videoWeb: null,
                },
              ],
              title: 'random',
            },
          ],
        },
        {
          date: '2020-11-25T00:00:00Z',
          albums: [
            {
              id: '523',
              title: 'another_album',
              media: [
                {
                  __typename: 'Media',
                  album: {
                    __typename: 'Album',
                    id: '523',
                    title: 'another_album',
                  },
                  date: '2020-11-25T16:14:33Z',
                  favorite: false,
                  highRes: {
                    __typename: 'MediaURL',
                    height: 4118,
                    url: 'http://localhost:4001/photo/122A2630-Edit_ySQWFAgE.jpg',
                    width: 6177,
                  },
                  id: '1059',
                  thumbnail: {
                    __typename: 'MediaURL',
                    height: 682,
                    url: 'http://localhost:4001/photo/thumbnail_122A2630-Edit_jpg_pwjtMkpy.jpg',
                    width: 1024,
                  },
                  title: '122A2630-Edit.jpg',
                  type: 'Photo',
                  videoWeb: null,
                },
                {
                  __typename: 'Media',
                  album: {
                    __typename: 'Album',
                    id: '523',
                    title: 'another_album',
                  },
                  date: '2020-11-25T16:43:59Z',
                  favorite: false,
                  highRes: {
                    __typename: 'MediaURL',
                    height: 884,
                    url: 'http://localhost:4001/photo/122A2785-2_mCnWjLdb.jpg',
                    width: 884,
                  },
                  id: '1060',
                  thumbnail: {
                    __typename: 'MediaURL',
                    height: 1024,
                    url: 'http://localhost:4001/photo/thumbnail_122A2785-2_jpg_CevmxEXf.jpg',
                    width: 1024,
                  },
                  title: '122A2785-2.jpg',
                  type: 'Photo',
                  videoWeb: null,
                },
              ],
            },
            {
              id: '522',
              title: 'random',
              media: [
                {
                  __typename: 'Media',
                  album: {
                    __typename: 'Album',
                    id: '522',
                    title: 'random',
                  },
                  date: '2020-11-25T16:14:33Z',
                  favorite: false,
                  highRes: {
                    __typename: 'MediaURL',
                    height: 4118,
                    url: 'http://localhost:4001/photo/122A2630-Edit_em9g89qg.jpg',
                    width: 6177,
                  },
                  id: '1056',
                  thumbnail: {
                    __typename: 'MediaURL',
                    height: 682,
                    url: 'http://localhost:4001/photo/thumbnail_122A2630-Edit_jpg_aJPCSDDl.jpg',
                    width: 1024,
                  },
                  title: '122A2630-Edit.jpg',
                  type: 'Photo',
                  videoWeb: null,
                },
              ],
            },
          ],
        },
        {
          date: '2020-11-09T00:00:00Z',
          albums: [
            {
              id: '522',
              title: 'random',
              media: [
                {
                  __typename: 'Media',
                  id: '1054',
                  title: '122A2559.jpg',
                  type: MediaType.Photo,
                  thumbnail: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/thumbnail_122A2559_jpg_MsOJtPi8.jpg',
                    width: 1024,
                    height: 712,
                  },
                  highRes: {
                    __typename: 'MediaURL',
                    url: 'http://localhost:4001/photo/122A2559_FDsQHuBN.jpg',
                    width: 6246,
                    height: 4346,
                  },
                  videoWeb: null,
                  favorite: false,
                  album: { __typename: 'Album', id: '522', title: 'random' },
                  date: '2020-11-09T15:38:09Z',
                },
              ],
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
          date: 1,
          album: 0,
          media: 0,
        },
      },
      {
        name: 'next album',
        in: {
          date: 1,
          album: 0,
          media: 1,
        },
        out: {
          date: 1,
          album: 1,
          media: 0,
        },
      },
      {
        name: 'next date',
        in: {
          date: 1,
          album: 1,
          media: 0,
        },
        out: {
          date: 2,
          album: 0,
          media: 0,
        },
      },
      {
        name: 'reached end',
        in: {
          date: 2,
          album: 0,
          media: 0,
        },
        out: {
          date: 2,
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
          date: 1,
          album: 1,
          media: 0,
        },
        out: {
          date: 1,
          album: 0,
          media: 1,
        },
      },
      {
        name: 'previous date',
        in: {
          date: 2,
          album: 0,
          media: 0,
        },
        out: {
          date: 1,
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
