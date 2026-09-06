import React from 'react'
import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import AlbumTree, { ALBUM_TREE_ROOT_QUERY } from './AlbumTree'
import { ALBUM_TREE_SUB_ALBUMS_QUERY } from './AlbumTreeNode'

const mocks = [
  {
    request: { query: ALBUM_TREE_ROOT_QUERY, variables: { showHidden: false } },
    result: {
      data: {
        myAlbums: [
          { id: '1', title: 'Root', viewerHidden: false, viewerIsOwner: true },
        ],
      },
    },
  },
  {
    request: {
      query: ALBUM_TREE_SUB_ALBUMS_QUERY,
      variables: { id: '1', showHidden: false },
    },
    result: {
      data: {
        album: {
          id: '1',
          subAlbums: [
            { id: '2', title: 'ChildA', viewerHidden: false },
            { id: '3', title: 'ChildB', viewerHidden: false },
          ],
        },
      },
    },
  },
]

// Rects keyed by the element's own title text, looked up via a
// getBoundingClientRect spy below (jsdom doesn't compute real layout).
type Rects = Record<string, DOMRect>

const mockRect = (rect: Partial<DOMRect>): DOMRect => ({
  top: 0,
  bottom: 0,
  left: 0,
  right: 0,
  width: 0,
  height: 0,
  x: 0,
  y: 0,
  toJSON: () => '',
  ...rect,
})

const stubRects = (rects: Rects) => {
  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(
    function (this: Element) {
      if (this.tagName === 'NAV') return rects.CONTAINER ?? mockRect({})
      if (this.tagName === 'LI') {
        const ownTitle = this.querySelector(':scope > div > a')?.textContent
        if (ownTitle && rects[ownTitle]) return rects[ownTitle]
      }
      return mockRect({})
    }
  )
}

afterEach(() => {
  vi.restoreAllMocks()
})

test('scrolls down to reveal the newly expanded children', async () => {
  stubRects({
    CONTAINER: mockRect({ top: 0, bottom: 200 }),
    Root: mockRect({ top: 150, bottom: 170 }),
    ChildB: mockRect({ top: 250, bottom: 270 }),
  })

  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        <AlbumTree />
      </MemoryRouter>
    </MockedProvider>
  )

  const nav = screen.getByRole('navigation')
  expect(nav.scrollTop).toBe(0)

  await userEvent.click(await screen.findByLabelText('Expand album'))
  await screen.findByText('ChildB')

  // Container bottom (200) sits 70px above the last child's bottom (270),
  // and the expanded node (bottom 170) stays comfortably in view after
  // scrolling by that amount, so the full 70px scroll should apply.
  expect(nav.scrollTop).toBe(70)
})

test('pins the expanded node to the top instead of hiding it', async () => {
  stubRects({
    CONTAINER: mockRect({ top: 0, bottom: 100 }),
    Root: mockRect({ top: 10, bottom: 30 }),
    ChildB: mockRect({ top: 300, bottom: 320 }),
  })

  render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        <AlbumTree />
      </MemoryRouter>
    </MockedProvider>
  )

  const nav = screen.getByRole('navigation')

  await userEvent.click(await screen.findByLabelText('Expand album'))
  await screen.findByText('ChildB')

  // Scrolling far enough to reveal ChildB (220px) would push Root's row
  // (top 10) above the container entirely, so it should scroll just
  // enough (10px) to pin Root to the top instead.
  expect(nav.scrollTop).toBe(10)
})
