import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import TimelineGallery, { MY_TIMELINE_QUERY } from './TimelineGallery'
import { timelineData } from './timelineTestData'

vi.mock('../../hooks/useScrollPagination')

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
  ]

  render(
    <MemoryRouter initialEntries={['/timeline']}>
      <MockedProvider mocks={graphqlMocks}>
        <TimelineGallery />
      </MockedProvider>
    </MemoryRouter>
  )

  expect(screen.queryByLabelText('Show only favorites')).toBeInTheDocument()

  expect(await screen.findAllByRole('link')).toHaveLength(4)
  expect(await screen.findAllByRole('img')).toHaveLength(5)
})
