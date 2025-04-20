import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event';
import TimelineGallery, { MY_TIMELINE_QUERY } from './TimelineGallery'
import { timelineData } from './timelineTestData'
import { renderWithProviders } from '../../helpers/testUtils'
import { gql } from '@apollo/client'

vi.mock('../../hooks/useScrollPagination')

// Define the missing query that's used by TimelineFilters component
const EARLIEST_MEDIA_QUERY = gql`
  query earliestMedia {
    myMedia(
      order: { order_by: "date_shot", order_direction: ASC }
      paginate: { limit: 1 }
    ) {
      id
      date
    }
  }
`

test('timeline with media', async () => {
  const graphqlMocks = [
    {
      request: {
        query: MY_TIMELINE_QUERY,
        variables: { onlyFavorites: false, offset: 0, limit: 200 },
      },
      result: {
        data: {
          myTimeline: timelineData,
        },
      },
    },
    {
      request: {
        query: EARLIEST_MEDIA_QUERY,
        variables: {},
      },
      result: {
        data: {
          myMedia: [
            {
              id: '1001',
              date: '2020-01-01T00:00:00Z',
            }
          ]
        }
      }
    }
  ]

  renderWithProviders(<TimelineGallery />, {
    mocks: graphqlMocks,
    initialEntries: ['/timeline']
  })

  expect(screen.getByLabelText('Show only favorites')).toBeInTheDocument()
  expect(await screen.findAllByRole('link')).toHaveLength(4)
  expect(await screen.findAllByRole('img')).toHaveLength(5)
})

test('shows loading state', async () => {
  const earliestMediaMock = {
    request: {
      query: EARLIEST_MEDIA_QUERY,
      variables: {},
    },
    result: {
      data: {
        myMedia: [{ id: '1001', date: '2020-01-01T00:00:00Z' }]
      }
    },
  };

  const timelineMock = {
    request: {
      query: MY_TIMELINE_QUERY,
      variables: { onlyFavorites: false, offset: 0, limit: 200 },
    },
    result: { data: { myTimeline: timelineData } },
    delay: 200 // Delay to ensure we catch the loading state
  };

  renderWithProviders(<TimelineGallery />, {
    mocks: [earliestMediaMock, timelineMock],
    initialEntries: ['/timeline']
  });

  // During loading, the favorites checkbox exists but no images yet
  expect(screen.getByLabelText('Show only favorites')).toBeInTheDocument();
  expect(screen.queryAllByRole('img')).toHaveLength(0);

  // After loading completes, images should appear
  expect(await screen.findAllByRole('img')).toHaveLength(5);
})

test('filter by favorites', async () => {
  // Create filtered data with known favorite items
  const favoriteTimelineData = timelineData.filter(item => item.favorite);

  // Make sure we have at least one favorite item in test data
  if (favoriteTimelineData.length === 0) {
    // Create a copy with at least one favorite item if needed
    favoriteTimelineData.push({ ...timelineData[0], favorite: true });
  }

  // Setup user event simulation separately
  const user = userEvent.setup();

  // Setup mocks - CRITICAL: Match exact variable ordering from error message
  const mocks = [
    // EARLIEST_MEDIA_QUERY mock
    {
      request: {
        query: EARLIEST_MEDIA_QUERY,
        variables: {},
      },
      result: {
        data: {
          myMedia: [{ id: '1001', date: '2020-01-01T00:00:00Z' }]
        }
      },
    },
    // Initial view (all items) - match exact variable order from network request
    {
      request: {
        query: MY_TIMELINE_QUERY,
        variables: { onlyFavorites: false, fromDate: undefined, offset: 0, limit: 200 },
      },
      result: { data: { myTimeline: timelineData } },
    },
    // Favorites-only view - match exact variable order from network request
    {
      request: {
        query: MY_TIMELINE_QUERY,
        variables: { onlyFavorites: true, fromDate: undefined, offset: 0, limit: 200 },
      },
      result: { data: { myTimeline: favoriteTimelineData } },
    },
  ];

  renderWithProviders(<TimelineGallery />, {
    mocks,
    initialEntries: ['/timeline']
  });

  // Wait for initial data to load
  await waitFor(() => {
    expect(screen.queryAllByRole('img').length).toBeGreaterThan(0);
  }, { timeout: 2000 });

  // Toggle favorites filter
  const checkbox = screen.getByLabelText('Show only favorites');
  await user.click(checkbox);

  // Wait for filtered data to load
  await waitFor(() => {
    // Check URL parameter was updated first, as this happens immediately
    expect(window.location.search).toContain('favorites=1');
  }, { timeout: 1000 });

  // Longer timeout for image loading
  await waitFor(() => {
    const images = screen.getAllByRole('img');
    expect(images.length).toBe(favoriteTimelineData.length);
  }, { timeout: 3000 });
})
