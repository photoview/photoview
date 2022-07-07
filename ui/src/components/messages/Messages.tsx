import React, { useState } from 'react'
import styled from 'styled-components'
import { authToken } from '../../helpers/authentication'
import MessageProgress from './MessageProgress'
import MessagePlain from './Message'
import SubscriptionsHook, { Message } from './SubscriptionsHook'
import { NotificationType } from '../../__generated__/globalTypes'

const Container = styled.div`
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 500px;

  @media (max-width: 1000px) {
    display: none;
  }
`

type MessageStateType = {
  set: React.Dispatch<React.SetStateAction<Message[]>>
  get: Message[]
  add(message: Message): void
  removeKey(key: string): void
}

export const MessageState: MessageStateType = {
  set: fn => {
    console.warn('set function is not defined yet, called with', fn)
  },
  get: [],
  add: message => {
    MessageState.set(messages => {
      const newMessages = messages.filter(msg => msg.key != message.key)
      newMessages.push(message)

      return newMessages
    })
  },
  removeKey: key => {
    MessageState.set(messages => {
      const newMessages = messages.filter(msg => msg.key != key)
      return newMessages
    })
  },
}

const Messages = () => {
  const [messages, setMessages] = useState<Message[]>([])
  MessageState.set = setMessages
  MessageState.get = messages

  const getMessageElement = (message: Message): React.FunctionComponent => {
    const dismissMessage = (message: Message) => {
      message.onDismiss && message.onDismiss()
      setMessages(messages => messages.filter(msg => msg.key != message.key))
    }

    switch (message.type) {
      case NotificationType.Message:
        return props => (
          <MessagePlain
            onDismiss={() => {
              dismissMessage(message)
            }}
            {...message.props}
            {...props}
          />
        )
      case NotificationType.Progress:
        return props => (
          <MessageProgress
            onDismiss={() => {
              dismissMessage(message)
            }}
            {...message.props}
            {...props}
          />
        )
      default:
        throw new Error(`Invalid message type: ${message.type}`)
    }
  }

  // const transitions = useTransition(messages.slice().reverse(), x => x.key, {
  //   from: {
  //     opacity: 0,
  //     height: '0px',
  //   },
  //   enter: {
  //     opacity: 1,
  //     height: `100px`,
  //   },
  //   leave: { opacity: 0, height: '0px' },
  // })

  const messageElems = messages.map(msg => {
    const Elem = getMessageElement(msg)
    return (
      <div key={msg.key}>
        <Elem />
      </div>
    )
  })

  return (
    <Container>
      {messageElems}
      {authToken() && (
        <SubscriptionsHook messages={messages} setMessages={setMessages} />
      )}
    </Container>
  )
}

export default Messages
