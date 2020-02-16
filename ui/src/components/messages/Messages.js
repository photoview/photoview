import React, { useState, useEffect } from 'react'
import { useTransition, animated } from 'react-spring'
import styled from 'styled-components'
import { Message } from 'semantic-ui-react'

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
    MessageState.set(messages => [...messages, message])
  },
}

const Messages = () => {
  if (!localStorage.getItem('token')) {
    return null
  }

  console.log('Rendering messages')

  const [messages, setMessages] = useState([])
  MessageState.set = setMessages
  MessageState.get = messages

  const [refMap] = useState(() => new WeakMap())

  const getMessageElement = (message, ref) => {
    const dismissMessage = key => {
      setMessages(messages => messages.filter(msg => msg.key != key))
    }

    const RefDiv = props => <div {...props} ref={x => x && ref(x)} />

    switch (message.type) {
      case 'message':
        return props => (
          <Message
            as={RefDiv}
            onDismiss={() => {
              dismissMessage(message.key)
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
              dismissMessage(message.key)
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

    console.log(resolveFunc, waitPromise)
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
      console.log('HERE', refMap, item)

      const refPromise = refHooks.get(item.key).promise
      console.log('promise', refPromise)

      await refPromise
      console.log('AFTER PROMISE', refMap, item)

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
          console.log('GET REF', refMap, refHooks, item.key)
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
      <SubscriptionsHook messages={messages} setMessages={setMessages} />
    </Container>
  )
}

export default Messages
