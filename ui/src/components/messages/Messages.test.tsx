import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { NotificationType } from '../../__generated__/globalTypes'
import { MessageProvider } from './MessageState'
import MessagePlain from './Message'
import MessageProgress from './MessageProgress'
import { Message } from './SubscriptionsHook'

interface MockSubscriptionsHookProps {
  messages: Message[]
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>
}

// Define the mock for SubscriptionsHook before using it
const MockSubscriptionsHook = ({ messages, setMessages }: MockSubscriptionsHookProps) => {
  return (
    <div>
      <button
        type="button"
        onClick={() => setMessages((prev: Message[]) => [...prev, ...messages])}
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

const PROGRESS_COLORS = {
  LOW: '#dc2625',    // red for < 30%
  MEDIUM: '#fbbf24', // yellow for 30-70%
  HIGH: '#56e263'    // green for > 70%
}

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
      <MessageProvider>
        <MessagePlain {...messages[0].props} />
        <MessagePlain {...messages[1].props} />
        <MessagePlain {...messages[2].props} />
      </MessageProvider>
    )

    const neutralMessage = screen.getByText('This is a neutral message.').closest('div')
    const positiveMessage = screen.getByText('This is a positive message.').closest('div')
    const negativeMessage = screen.getByText('This is a negative message.').closest('div')

    expect(neutralMessage?.parentElement).toHaveClass('bg-white')
    expect(neutralMessage?.parentElement).toHaveClass('dark:bg-dark-bg2')
    expect(positiveMessage?.parentElement).toHaveClass('bg-green-100')
    expect(positiveMessage?.parentElement).toHaveClass('dark:bg-green-900')
    expect(negativeMessage?.parentElement).toHaveClass('bg-red-100')
    expect(negativeMessage?.parentElement).toHaveClass('dark:bg-red-900')
  })

  test('renders messages with overflow and scroll behavior', () => {
    const longContent = 'This is a very long message '.repeat(10).trim()

    render(
      <MessageProvider>
        <MessagePlain
          {...messages[0].props}
          content={longContent}
        />
      </MessageProvider>
    )

    const longMessage = screen.getByText((content, element) => {
      return element?.textContent === longContent
    })
    expect(longMessage.closest('div')).toHaveStyle({ maxHeight: '6rem' })
    expect(longMessage.closest('div')).toHaveStyle({ overflowY: 'auto' })
  })

  test('renders progress messages with correct progress bar color and width', () => {
    const { container } = render(
      <MessageProvider>
        <MessageProgress {...messages[3].props} />
      </MessageProvider>
    )

    const progressBar = container.querySelector('div[role="progressbar"][style*="width: 50%"]')
    expect(progressBar).toBeInTheDocument()
    expect(progressBar).toHaveStyle(`background-color: ${PROGRESS_COLORS.MEDIUM}`)
  })

  test('renders progress bar with correct colors based on percent value', () => {
    render(
      <MessageProvider>
        <MessageProgress {...messages[3].props} percent={20} />
        <MessageProgress {...messages[3].props} percent={50} />
        <MessageProgress {...messages[3].props} percent={80} />
      </MessageProvider>
    )

    const progressBars = screen.getAllByRole('progressbar')

    expect(progressBars[0]).toHaveStyle(`background-color: ${PROGRESS_COLORS.LOW}`)    // red for 20%
    expect(progressBars[1]).toHaveStyle(`background-color: ${PROGRESS_COLORS.MEDIUM}`) // yellow for 50%
    expect(progressBars[2]).toHaveStyle(`background-color: ${PROGRESS_COLORS.HIGH}`)   // green for 80%
  })

  test('subscriptions hook triggers messages correctly', () => {
    const setMessages = vi.fn()

    render(
      <MessageProvider>
        <MockSubscriptionsHook messages={messages} setMessages={setMessages} />
      </MessageProvider>
    )

    fireEvent.click(screen.getByTestId('trigger-messages'))
    expect(setMessages).toHaveBeenCalledWith(expect.any(Function))
  })
})
