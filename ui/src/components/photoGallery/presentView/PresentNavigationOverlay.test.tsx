import React from 'react'
import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import PresentNavigationOverlay from './PresentNavigationOverlay'

const hidden = (label: string) =>
  screen.getByLabelText(label).classList.contains('hide')

beforeEach(() => {
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

test('a key press reveals the controls and they hide themselves again', () => {
  render(<PresentNavigationOverlay dispatchMedia={vi.fn()} />)

  expect(hidden('Exit presentation mode')).toBe(true)

  // A keyboard user has no pointer to move, and hidden controls are out of the
  // tab order, so any key has to be a way in.
  fireEvent.keyDown(document, { key: 'ArrowRight' })
  expect(hidden('Exit presentation mode')).toBe(false)

  // The single press schedules the hide as well - the debounce's trailing edge
  // always runs, so the reveal is not a state the controls get stuck in.
  act(() => {
    vi.advanceTimersByTime(2000)
  })
  expect(hidden('Exit presentation mode')).toBe(true)

  // And the next press reveals them again rather than toggling back the other
  // way: rising and trailing edges alternate, so the phase cannot drift.
  fireEvent.keyDown(document, { key: 'ArrowRight' })
  expect(hidden('Exit presentation mode')).toBe(false)
})

test('a tap reveals the controls and repeated taps keep them visible', () => {
  render(<PresentNavigationOverlay dispatchMedia={vi.fn()} />)

  fireEvent.click(screen.getByTestId('present-overlay'))
  expect(hidden('Next image')).toBe(false)

  // Tapping again inside the window only pushes the auto-hide back; it must
  // not hide the controls under the finger that is reaching for them.
  act(() => {
    vi.advanceTimersByTime(1500)
  })
  fireEvent.click(screen.getByTestId('present-overlay'))
  act(() => {
    vi.advanceTimersByTime(1500)
  })
  expect(hidden('Next image')).toBe(false)

  act(() => {
    vi.advanceTimersByTime(500)
  })
  expect(hidden('Next image')).toBe(true)
})

test('a horizontal swipe navigates between images', () => {
  vi.useRealTimers()

  const dispatchMedia = vi.fn()
  render(<PresentNavigationOverlay dispatchMedia={dispatchMedia} />)

  const surface = screen.getByTestId('present-overlay').firstElementChild
  if (surface == null) throw new Error('swipe surface not found')

  const swipe = (from: number, to: number) => {
    fireEvent.touchStart(surface, {
      touches: [{ clientX: from, clientY: 100 }],
    })
    fireEvent.touchMove(surface, { touches: [{ clientX: to, clientY: 100 }] })
    fireEvent.touchEnd(surface, {
      changedTouches: [{ clientX: to, clientY: 100 }],
    })
  }

  swipe(300, 100)
  swipe(100, 300)

  expect(dispatchMedia).toHaveBeenNthCalledWith(1, { type: 'nextImage' })
  expect(dispatchMedia).toHaveBeenNthCalledWith(2, { type: 'previousImage' })
})

test('the exit button leaves the viewer and steps back in history', async () => {
  vi.useRealTimers()

  const dispatchMedia = vi.fn()
  const back = vi
    .spyOn(window.history, 'back')
    .mockImplementation(() => undefined)

  render(<PresentNavigationOverlay dispatchMedia={dispatchMedia} />)

  // Hidden controls take no clicks, so the first tap is the one that reveals
  // them - the same two steps a touch user takes.
  fireEvent.click(screen.getByTestId('present-overlay'))
  await userEvent.click(screen.getByLabelText('Exit presentation mode'))

  expect(dispatchMedia).toHaveBeenCalledWith({ type: 'closePresentMode' })
  expect(
    back,
    'leaving restores the URL the viewer came from'
  ).toHaveBeenCalled()

  back.mockRestore()
})

test('the exit button leaves history alone when the caller asks it to', async () => {
  vi.useRealTimers()

  const dispatchMedia = vi.fn()
  const back = vi
    .spyOn(window.history, 'back')
    .mockImplementation(() => undefined)

  render(
    <PresentNavigationOverlay
      dispatchMedia={dispatchMedia}
      disableSaveCloseInHistory
    />
  )

  fireEvent.click(screen.getByTestId('present-overlay'))
  await userEvent.click(screen.getByLabelText('Exit presentation mode'))

  expect(dispatchMedia).toHaveBeenCalledWith({ type: 'closePresentMode' })
  expect(back, 'this caller manages history itself').not.toHaveBeenCalled()

  back.mockRestore()
})
