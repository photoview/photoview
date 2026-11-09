import { NotificationType } from '../../__generated__/globalTypes'
import { Message, withSubscriptionError } from './SubscriptionsHook'

const unrelated: Message = {
  key: 'unrelated',
  type: NotificationType.Message,
  props: { header: 'Something else', content: '' },
}

test('a subscription that keeps failing shows one message, not one per attempt', () => {
  // Every reconnect of a stream whose server stays down reports a new error.
  let messages = [unrelated]
  for (const failure of ['socket closed', 'socket closed again', 'and again']) {
    messages = withSubscriptionError(messages, new Error(failure))
  }

  const errors = messages.filter(m => m.props.header === 'Network error')
  expect(errors).toHaveLength(1)
  // The newest failure is the one on screen.
  expect(errors[0].props.content).toBe('and again')
  // And other messages are left alone.
  expect(messages).toContain(unrelated)
})
