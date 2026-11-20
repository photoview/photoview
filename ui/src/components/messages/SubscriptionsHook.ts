import { notificationSubscription } from './__generated__/notificationSubscription'
import { useEffect } from 'react'
import i18n from 'i18next'
import { useSubscription, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { NotificationType } from '../../__generated__/globalTypes'

const NOTIFICATION_SUBSCRIPTION = gql`
  subscription notificationSubscription {
    notification {
      key
      type
      header
      content
      progress
      positive
      negative
      timeout
    }
  }
`

const messageTimeoutHandles = new Map<string, number>()

export interface Message {
  key: string
  type: NotificationType
  timeout?: number
  onDismiss?: () => void
  props: {
    header: string
    content: string
    negative?: boolean
    positive?: boolean
    percent?: number
  }
}

// One key for every error of this subscription. A server that stays down
// fails the reconnecting stream again on every attempt, and each failure
// replacing the previous one keeps that to a single message.
const SUBSCRIPTION_ERROR_KEY = 'notification-subscription-error'

// Exported for its tests.
export const withSubscriptionError = (
  messages: Message[],
  error: Error
): Message[] => [
  ...messages.filter(m => m.key != SUBSCRIPTION_ERROR_KEY),
  {
    key: SUBSCRIPTION_ERROR_KEY,
    type: NotificationType.Message,
    props: {
      header: i18n.t('notification.error.network.header', 'Network error'),
      content: error.message,
      negative: true,
    },
  },
]

type SubscriptionHookProps = {
  messages: Message[]
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>
}

const SubscriptionsHook = ({
  messages,
  setMessages,
}: SubscriptionHookProps) => {
  if (!authToken()) {
    return null
  }

  const { data, error } = useSubscription<notificationSubscription>(
    NOTIFICATION_SUBSCRIPTION
  )

  useEffect(() => {
    if (error) {
      setMessages(state => withSubscriptionError(state, error))
    }

    if (!data) return

    const newMessages = [...messages]

    const msg = data.notification

    if (msg.type == 'Close') {
      setMessages(messages => messages.filter(m => m.key != msg.key))
      return
    }

    const newNotification: Message = {
      key: msg.key,
      type: msg.type,
      timeout: msg.timeout || undefined,
      props: {
        header: msg.header,
        content: msg.content,
        negative: msg.negative,
        positive: msg.positive,
        percent: msg.progress || undefined,
      },
    }

    if (msg.timeout) {
      // Clear old timeout, to replace it with the new one
      if (messageTimeoutHandles.get(msg.key)) {
        const timeoutHandle = messageTimeoutHandles.get(msg.key)
        clearTimeout(timeoutHandle)
      }

      const timeoutHandle = setTimeout(() => {
        setMessages(messages => messages.filter(m => m.key != msg.key))
      }, msg.timeout) as unknown as number

      messageTimeoutHandles.set(msg.key, timeoutHandle)
    }

    const notifyIndex = newMessages.findIndex(
      msg => msg.key == newNotification.key
    )
    if (notifyIndex != -1) {
      newMessages[notifyIndex] = newNotification
    } else {
      newMessages.push(newNotification)
    }

    setMessages(newMessages)
  }, [data, error])

  return null
}

export default SubscriptionsHook
