import { screen } from '@testing-library/react'
import MediaSidebar, {
  MediaSidebarMedia,
  SIDEBAR_MEDIA_QUERY
} from './MediaSidebar'
import { SIDEBAR_DOWNLOAD_QUERY } from '../SidebarDownloadMedia'
import { MediaType } from '../../../__generated__/globalTypes'
import { renderWithProviders } from '../../../helpers/testUtils'
import { gql } from '@apollo/client'
import * as authentication from '../../../helpers/authentication'

vi.mock('../../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

// Define the photo shares query directly in the test file
const SIDEBAR_GET_PHOTO_SHARES = gql`
  query sidebarGetPhotoShares($id: ID!) {
    media(id: $id) {
      id
      shares {
        id
        token
        hasPassword
      }
    }
  }
`

describe('MediaSidebar', () => {
  const media: MediaSidebarMedia = {
    __typename: 'Media',
    id: '6867',
    title: '122A6069.jpg',
    type: MediaType.Photo,
    thumbnail: {
      __typename: 'MediaURL',
      url: '/photo/thumbnail.jpg',
      width: 1024,
      height: 839,
    },
    highRes: {
      __typename: 'MediaURL',
      url: '/photo/highres.jpg',
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

  // Create mocks for all required GraphQL queries
  const mocks = [
    {
      request: {
        query: SIDEBAR_DOWNLOAD_QUERY,
        variables: { mediaId: '6867' }
      },
      result: {
        data: {
          media: {
            __typename: 'Media',
            id: '6867',
            downloads: [
              // Include at least one properly structured download item
              {
                __typename: 'Download',
                title: 'Original',
                mediaUrl: {
                  __typename: 'MediaURL',
                  url: '/download/original.jpg',
                  width: 5322,
                  height: 4362,
                  fileSize: 1234567
                }
              }
            ]
          }
        }
      }
    },
    {
      request: {
        query: SIDEBAR_GET_PHOTO_SHARES,
        variables: { id: '6867' }
      },
      result: {
        data: {
          media: {
            __typename: 'Media',
            id: '6867',
            shares: []
          }
        }
      }
    },
    {
      request: {
        query: SIDEBAR_MEDIA_QUERY,
        variables: { id: '6867' }
      },
      result: {
        data: {
          media: {
            __typename: 'Media',
            id: '6867',
            title: '122A6069.jpg',
            type: MediaType.Photo,
            highRes: {
              __typename: 'MediaURL',
              url: '/photo/highres.jpg',
              width: 5322,
              height: 4362,
            },
            thumbnail: {
              __typename: 'MediaURL',
              url: '/photo/thumbnail.jpg',
              width: 1024,
              height: 839,
            },
            videoWeb: null,
            videoMetadata: null,
            exif: null,
            album: {
              __typename: 'Album',
              id: '2294',
              title: 'album_name',
              path: []
            },
            faces: []
          }
        }
      }
    }
  ]

  test('render sample image, unauthorized', () => {
    authToken.mockImplementation(() => null)

    // Only need the download query for unauthorized view
    renderWithProviders(<MediaSidebar media={media} />, {
      mocks: [mocks[0]],
      apolloOptions: {
        addTypename: true,
        defaultOptions: {
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' }
        }
      }
    })

    expect(screen.getByText('122A6069.jpg')).toBeInTheDocument()
    expect(screen.getByRole('img')).toHaveAttribute(
      'src',
      'http://localhost:3000/photo/highres.jpg'
    )

    expect(
      screen.queryByText('Set as album cover photo')
    ).not.toBeInTheDocument()
    expect(screen.queryByText('Sharing options')).not.toBeInTheDocument()
  })

  test('render sample image, authorized', () => {
    authToken.mockImplementation(() => 'token-here')

    // Need all mocks for authorized view
    renderWithProviders(<MediaSidebar media={media} />, {
      mocks: mocks,
      apolloOptions: {
        addTypename: true,
        defaultOptions: {
          watchQuery: { fetchPolicy: 'no-cache' },
          query: { fetchPolicy: 'no-cache' }
        }
      }
    })

    expect(screen.getByText('122A6069.jpg')).toBeInTheDocument()
    expect(screen.getByRole('img')).toHaveAttribute(
      'src',
      'http://localhost:3000/photo/highres.jpg'
    )

    expect(screen.getByText('Set as album cover photo')).toBeInTheDocument()
    expect(screen.getByText('Album path')).toBeInTheDocument()
  })

  test('displays loading state correctly', () => {
    // Use the media object already defined in the describe block
    authToken.mockImplementation(() => 'token-here')

    // Mock loadMedia to show loading state
    const loadMediaMock = vi.fn()
    vi.spyOn(require('@apollo/client'), 'useLazyQuery').mockReturnValue([
      loadMediaMock,
      { loading: true, error: undefined, data: null }
    ])

    renderWithProviders(<MediaSidebar media={media} />)

    // Should show the media from props while loading
    expect(screen.getByText('122A6069.jpg')).toBeInTheDocument()
  })

  test('displays error state correctly', () => {
    // Use the media object already defined in the describe block
    authToken.mockImplementation(() => 'token-here')

    // Mock a GraphQL error
    vi.spyOn(require('@apollo/client'), 'useLazyQuery').mockReturnValue([
      vi.fn(),
      { loading: false, error: new Error('Failed to load media'), data: null }
    ])

    renderWithProviders(<MediaSidebar media={media} />)

    // Should show the error message
    expect(screen.getByText(/Failed to load media/)).toBeInTheDocument()
  })

  test('renders video content correctly', () => {
    // Create a video variant of the media object
    const videoMedia: MediaSidebarMedia = {
      ...media,
      type: MediaType.Video,
      videoWeb: {
        __typename: 'MediaURL',
        url: '/video/web.mp4',
        width: 1280,
        height: 720
      }
    }

    renderWithProviders(<MediaSidebar media={videoMedia} />)

    // Should render video element instead of image
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
    // Video testing would depend on how your ProtectedVideo component renders
  })
})
