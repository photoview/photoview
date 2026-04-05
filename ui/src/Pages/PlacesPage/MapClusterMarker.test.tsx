import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MapClusterMarker from './MapClusterMarker'
import { MediaMarker } from './MapPresentMarker'

const clusterMarker: MediaMarker = {
  id: 1,
  thumbnail: JSON.stringify({ url: 'http://example.com/thumb.jpg' }),
  cluster: true,
  point_count_abbreviated: 5,
  cluster_id: 'c42',
  media_id: 'm1',
}

const singleMarker: MediaMarker = {
  id: 2,
  thumbnail: JSON.stringify({ url: 'http://example.com/single.jpg' }),
  cluster: false,
  point_count_abbreviated: 0,
  cluster_id: '',
  media_id: 'm7',
}

describe('MapClusterMarker', () => {
  test('renders thumbnail image from parsed marker thumbnail', () => {
    const dispatch = vi.fn()
    render(
      <MapClusterMarker marker={clusterMarker} dispatchMarkerMedia={dispatch} />
    )
    const imgs = screen.getAllByRole('img')
    const thumbnail = imgs.find(
      img => img.getAttribute('src') === 'http://example.com/thumb.jpg'
    )
    expect(thumbnail).toBeDefined()
  })

  test('renders point count badge for cluster markers', () => {
    const dispatch = vi.fn()
    render(
      <MapClusterMarker marker={clusterMarker} dispatchMarkerMedia={dispatch} />
    )
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  test('does not render point count badge for non-cluster markers', () => {
    const dispatch = vi.fn()
    render(
      <MapClusterMarker marker={singleMarker} dispatchMarkerMedia={dispatch} />
    )
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  test('dispatches replacePresentMarker with cluster info on click', async () => {
    const dispatch = vi.fn()
    const { container } = render(
      <MapClusterMarker marker={clusterMarker} dispatchMarkerMedia={dispatch} />
    )
    await userEvent.click(container.firstElementChild!)
    expect(dispatch).toHaveBeenCalledWith({
      type: 'replacePresentMarker',
      marker: { cluster: true, id: 'c42' },
    })
  })

  test('dispatches replacePresentMarker with media_id for non-cluster marker', async () => {
    const dispatch = vi.fn()
    const { container } = render(
      <MapClusterMarker marker={singleMarker} dispatchMarkerMedia={dispatch} />
    )
    await userEvent.click(container.firstElementChild!)
    expect(dispatch).toHaveBeenCalledWith({
      type: 'replacePresentMarker',
      marker: { cluster: false, id: 'm7' },
    })
  })
})
