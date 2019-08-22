import React, { useState, useEffect } from 'react'
import { useSubscription } from 'react-apollo'
import { useTransition, animated } from 'react-spring'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Message } from 'semantic-ui-react'
import gql from 'graphql-tag'
import MessageProgress from './MessageProgress'

const syncSubscription = gql`
  subscription syncSubscription {
    scannerStatusUpdate {
      finished
      success
      message
      progress
    }
  }
`

const Container = styled.div`
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 500px;
`

const Messages = () => {
  if (!localStorage.getItem('token')) {
    return null
  }

  const { data, error } = useSubscription(syncSubscription)

  const [messages, setMessages] = useState([])
  const [refMap] = useState(() => new WeakMap())

  useEffect(() => {
    if (error) {
      setMessages(state => [
        ...state,
        {
          key: Math.random().toString(26),
          type: 'message',
          props: {
            header: 'Network error',
            content: error.message,
            negative: true,
          },
        },
      ])
    }

    if (!data) return

    const update = data.scannerStatusUpdate
    const newMessages = [...messages]

    if (update.success) {
      newMessages[0] = {
        key: 'primary',
        type: 'progress',
        props: {
          header: update.finished ? 'Synced' : 'Syncing',
          content: update.message,
          percent: update.progress,
          positive: update.finished,
        },
      }

      if (!update.finished) newMessages[0].props.onDismiss = null
    } else {
      const key = Math.random().toString(26)
      newMessages.push({
        key,
        type: 'message',
        props: {
          header: 'Sync error',
          content: update.message,
          negative: true,
        },
      })
    }

    setMessages(newMessages)
  }, [data, error])

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

  const transitions = useTransition(messages.slice().reverse(), x => x.key, {
    from: {
      opacity: 0,
      height: '0px',
    },
    enter: item => async next => {
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
        }
        const MessageElement = getMessageElement(item, getRef)

        style.padding = 0

        return (
          <animated.div key={key} style={style}>
            <MessageElement />
          </animated.div>
        )
      })}
    </Container>
  )
}

export default Messages
