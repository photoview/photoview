import React from 'react'
import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MediaType } from '../../../__generated__/globalTypes'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'
import { SidebarContext } from '../../sidebar/Sidebar'
import PresentView from './PresentView'

const makeMedia = (id: string): MediaGalleryFields => ({
  __typename: 'Media',
  id,
  type: MediaType.Photo,
  highRes: null,
  blurhash: null,
  videoWeb: null,
  favorite: false,
  thumbnail: {
    __typename: 'MediaURL',
    url: `/photo-${id}.jpg`,
    width: 300,
    height: 200,
  },
})

/** Renders PresentView with a stand-in sidebar whose content the test owns. */
const renderWithSidebar = (media: MediaGalleryFields) => {
  const updateSidebar = vi.fn()
  let content: React.ReactNode = null

  const view = render(
    <SidebarContext.Provider
      value={{
        updateSidebar,
        setPinned: vi.fn(),
        content,
        pinned: false,
      }}
    >
      <PresentView activeMedia={media} dispatchMedia={vi.fn()} />
    </SidebarContext.Provider>
  )

  const rerender = (next: MediaGalleryFields, nextContent: React.ReactNode) => {
    content = nextContent
    view.rerender(
      <SidebarContext.Provider
        value={{ updateSidebar, setPinned: vi.fn(), content, pinned: false }}
      >
        <PresentView activeMedia={next} dispatchMedia={vi.fn()} />
      </SidebarContext.Provider>
    )
  }

  return { updateSidebar, rerender }
}

test('the info panel follows the image the viewer is on', async () => {
  // Stable objects: PresentView rebuilds the panel when the active media
  // changes, and a fresh object with the same id is a change.
  const first = makeMedia('1')
  const second = makeMedia('2')

  const { updateSidebar, rerender } = renderWithSidebar(first)

  // The controls hide themselves until the viewer asks for them.
  fireEvent.click(screen.getByTestId('present-overlay'))
  await userEvent.click(screen.getByLabelText('Show media info'))

  expect(updateSidebar).toHaveBeenCalledTimes(1)

  // The real sidebar keeps what it was handed, so the context now has content.
  rerender(first, <div>sidebar</div>)
  expect(updateSidebar).toHaveBeenCalledTimes(1)

  // Navigating with the panel open has to rebuild it: it was made from the
  // image that was active when it opened, and would otherwise keep showing
  // that one while the viewer looks at another.
  rerender(second, <div>sidebar</div>)
  expect(updateSidebar).toHaveBeenCalledTimes(2)
})

test('an info panel closed elsewhere stays closed while navigating', async () => {
  const first = makeMedia('1')
  const second = makeMedia('2')

  const { updateSidebar, rerender } = renderWithSidebar(first)

  fireEvent.click(screen.getByTestId('present-overlay'))
  await userEvent.click(screen.getByLabelText('Show media info'))
  rerender(first, <div>sidebar</div>)
  expect(updateSidebar).toHaveBeenCalledTimes(1)

  // The sidebar has its own close button. Once it is gone, moving to the next
  // image must not bring it back.
  rerender(first, null)
  rerender(second, null)

  expect(updateSidebar).toHaveBeenCalledTimes(1)
})

test('arrow keys navigate and escape leaves the viewer', () => {
  const dispatchMedia = vi.fn()

  render(
    <SidebarContext.Provider
      value={{
        updateSidebar: vi.fn(),
        setPinned: vi.fn(),
        content: null,
        pinned: false,
      }}
    >
      <PresentView
        activeMedia={makeMedia('1')}
        dispatchMedia={dispatchMedia}
        disableSaveCloseInHistory
      />
    </SidebarContext.Provider>
  )

  act(() => {
    fireEvent.keyDown(document, { key: 'ArrowRight' })
    fireEvent.keyDown(document, { key: 'ArrowLeft' })
    fireEvent.keyDown(document, { key: 'Escape' })
  })

  expect(dispatchMedia).toHaveBeenNthCalledWith(1, { type: 'nextImage' })
  expect(dispatchMedia).toHaveBeenNthCalledWith(2, { type: 'previousImage' })
  expect(dispatchMedia).toHaveBeenNthCalledWith(3, { type: 'closePresentMode' })
})

test('escape without the history flag steps back instead', () => {
  const dispatchMedia = vi.fn()
  const back = vi
    .spyOn(window.history, 'back')
    .mockImplementation(() => undefined)

  render(
    <SidebarContext.Provider
      value={{
        updateSidebar: vi.fn(),
        setPinned: vi.fn(),
        content: null,
        pinned: false,
      }}
    >
      <PresentView activeMedia={makeMedia('1')} dispatchMedia={dispatchMedia} />
    </SidebarContext.Provider>
  )

  fireEvent.keyDown(document, { key: 'Escape' })

  expect(dispatchMedia).toHaveBeenCalledWith({ type: 'closePresentMode' })
  expect(back).toHaveBeenCalled()

  back.mockRestore()
})
