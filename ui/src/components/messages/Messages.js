import React, { useState, useEffect } from 'react'
import { useTransition, animated } from 'react-spring'
import styled from 'styled-components'
import { Message } from 'semantic-ui-react'
import { authToken } from '../../authentication'

import MessageProgress from './MessageProgress'
import SubscriptionsHook from './SubscriptionsHook'

const Container = styled.div`
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 500px;
`

export let MessageState = {
  set: null,
  get: null,
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
  const [messages, setMessages] = useState([])
  MessageState.set = setMessages
  MessageState.get = messages

  const [refMap] = useState(() => new WeakMap())

  const getMessageElement = (message, ref) => {
    const dismissMessage = message => {
      message.onDismiss && message.onDismiss()
      setMessages(messages => messages.filter(msg => msg.key != message.key))
    }

    const RefDiv = props => <div {...props} ref={x => x && ref(x)} />

    switch (message.type) {
      case 'message':
        return props => (
          <Message
            as={RefDiv}
            onDismiss={() => {
              dismissMessage(message)
            }}
            floating
            {...message.props}
            {...props}
          />
        )
      case 'progress':
        return props => (
          <MessageProgress
            as={RefDiv}
            onDismiss={() => {
              dismissMessage(message)
            }}
            {...message.props}
            {...props}
          />
        )
    }
  }

  let refHooks = new Map()
  messages.forEach(message => {
    let resolveFunc = null

    const waitPromise = new Promise((resolve, reject) => {
      resolveFunc = resolve
    })

    refHooks.set(message.key, {
      done: resolveFunc,
      promise: waitPromise,
    })
  })

  const transitions = useTransition(messages.slice().reverse(), x => x.key, {
    from: {
      opacity: 0,
      height: '0px',
    },
    enter: item => async next => {
      const refPromise = refHooks.get(item.key).promise
      await refPromise

      await next({
        opacity: 1,
        height: `${refMap.get(item).offsetHeight + 10}px`,
      })
    },
    leave: { opacity: 0, height: '0px' },
  })

  return (
    <Container>
      {transitions.map(({ item, props: style, key }) => {
        const getRef = ref => {
          refMap.set(item, ref)
          if (refHooks.has(item.key)) {
            refHooks.get(item.key).done()
          }
        }
        const MessageElement = getMessageElement(item, getRef)

        style.padding = 0

        return (
          <animated.div key={key} style={style}>
            <MessageElement />
          </animated.div>
        )
      })}
      {authToken() && (
        <SubscriptionsHook messages={messages} setMessages={setMessages} />
      )}
    </Container>
  )
}

export default Messages
