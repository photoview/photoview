import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { NotificationType } from '../../__generated__/globalTypes'
import { MessageState } from './MessageState'
import MessagePlain from './Message'
import MessageProgress from './MessageProgress'

// Define the mock for SubscriptionsHook before using it
const MockSubscriptionsHook = ({ messages, setMessages }: any) => {
  return (
    <div>
      <button
        onClick={() => setMessages([...messages, ...messages])}
        data-testid="trigger-messages"
      >
        Trigger Messages
      </button>
    </div>
  )
}

vi.mock('./SubscriptionsHook', () => ({
  __esModule: true,
  default: MockSubscriptionsHook,
}))

const messages = [
  {
    key: '1',
    type: NotificationType.Message,
    timeout: 5000,
    props: {
      header: 'Info Message',
      content: 'This is a neutral message.',
    },
  },
  {
    key: '2',
    type: NotificationType.Message,
    timeout: 5000,
    props: {
      header: 'Success Message',
      content: 'This is a positive message.',
      positive: true,
    },
  },
  {
    key: '3',
    type: NotificationType.Message,
    timeout: 5000,
    props: {
      header: 'Error Message',
      content: 'This is a negative message.',
      negative: true,
    },
  },
  {
    key: '4',
    type: NotificationType.Progress,
    timeout: 5000,
    props: {
      header: 'Progress Message',
      content: 'This is a progress message.',
      percent: 50,
    },
  },
]

describe('Messages Component', () => {
  test('renders different types of messages with correct background colors', () => {
    render(
      <MessageState.Provider value={{ messages, setMessages: () => {} }}>
        <MessagePlain {...messages[0].props} />
        <MessagePlain {...messages[1].props} />
        <MessagePlain {...messages[2].props} />
      </MessageState.Provider>
    )

    const neutralMessage = screen.getByText('This is a neutral message.')
    const positiveMessage = screen.getByText('This is a positive message.')
    const negativeMessage = screen.getByText('This is a negative message.')

    expect(neutralMessage.closest('div')).toHaveClass('bg-white')
    expect(neutralMessage.closest('div')).toHaveClass('dark:bg-dark-bg2')
    expect(positiveMessage.closest('div')).toHaveClass('bg-green-100')
    expect(positiveMessage.closest('div')).toHaveClass('dark:bg-green-900')
    expect(negativeMessage.closest('div')).toHaveClass('bg-red-100')
    expect(negativeMessage.closest('div')).toHaveClass('dark:bg-red-900')
  })

  test('renders messages with overflow and scroll behavior', () => {
    const longContent = 'This is a very long message '.repeat(10)

    render(
      <MessageState.Provider value={{ messages, setMessages: () => {} }}>
        <MessagePlain
          {...messages[0].props}
          content={longContent}
        />
      </MessageState.Provider>
    )

    const longMessage = screen.getByText(longContent)
    expect(longMessage.closest('div')).toHaveStyle({ maxHeight: '6rem' })
    expect(longMessage.closest('div')).toHaveStyle({ overflowY: 'auto' })
  })

  test('renders progress messages with correct progress bar color and width', () => {
    const { container } = render(
      <MessageState.Provider value={{ messages, setMessages: () => {} }}>
        <MessageProgress {...messages[3].props} />
      </MessageState.Provider>
    )

    const progressBar = container.querySelector('div[style*="width: 50%"]')
    expect(progressBar).toBeInTheDocument()
    expect(progressBar).toHaveStyle('background-color: #fbbf24')
  })

  test('renders progress bar with correct colors based on percent value', () => {
    render(
      <MessageState.Provider value={{ messages, setMessages: () => {} }}>
        <MessageProgress {...messages[3].props} percent={20} />
        <MessageProgress {...messages[3].props} percent={50} />
        <MessageProgress {...messages[3].props} percent={80} />
      </MessageState.Provider>
    )

    const progressBars = screen.getAllByRole('progressbar')

    expect(progressBars[0]).toHaveStyle('background-color: #dc2625') // red for 20%
    expect(progressBars[1]).toHaveStyle('background-color: #fbbf24') // yellow for 50%
    expect(progressBars[2]).toHaveStyle('background-color: #56e263') // green for 80%
  })

  test('subscriptions hook triggers messages correctly', () => {
    const setMessages = vi.fn()

    render(
      <MessageState.Provider value={{ messages: [], setMessages }}>
        <MockSubscriptionsHook />
      </MessageState.Provider>
    )

    fireEvent.click(screen.getByTestId('trigger-messages'))
    expect(setMessages).toHaveBeenCalledWith([...messages, ...messages])
  })
})
