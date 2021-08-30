import React from 'react'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import { fireEvent, render, screen, act } from '@testing-library/react'

jest.useFakeTimers('modern')

describe('PresentNavigationOverlay component', () => {
  test('simple render', () => {
    const dispatchMedia = jest.fn()
    render(<PresentNavigationOverlay dispatchMedia={dispatchMedia} />)

    expect(screen.getByLabelText('Previous image')).toBeInTheDocument()
    expect(screen.getByLabelText('Next image')).toBeInTheDocument()
    expect(screen.getByLabelText('Exit presentation mode')).toBeInTheDocument()
  })

  test('click buttons', () => {
    const dispatchMedia = jest.fn()
    render(<PresentNavigationOverlay dispatchMedia={dispatchMedia} />)

    expect(dispatchMedia).not.toHaveBeenCalled()

    fireEvent.click(screen.getByLabelText('Next image'))
    expect(dispatchMedia).lastCalledWith({ type: 'nextImage' })

    fireEvent.click(screen.getByLabelText('Previous image'))
    expect(dispatchMedia).lastCalledWith({ type: 'previousImage' })
  })

  test('mouse move, show and hide', () => {
    const dispatchMedia = jest.fn()
    const { container } = render(
      <PresentNavigationOverlay dispatchMedia={dispatchMedia} />
    )

    expect(screen.getByLabelText('Next image')).toHaveClass('hide')

    fireEvent.mouseMove(container.firstChild!)
    expect(screen.getByLabelText('Next image')).not.toHaveClass('hide')

    act(() => {
      jest.advanceTimersByTime(3000)
    })

    expect(screen.getByLabelText('Next image')).toHaveClass('hide')
  })
})
