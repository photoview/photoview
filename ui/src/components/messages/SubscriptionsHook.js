import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import { useSubscription } from 'react-apollo'
import gql from 'graphql-tag'

const notificationSubscription = gql`
  subscription notificationSubscription {
    notification {
      key
      type
      header
      content
      progress
      positive
      negative
    }
  }
`

const SubscriptionsHook = ({ messages, setMessages }) => {
  const { data, error } = useSubscription(notificationSubscription)

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

    const newMessages = [...messages]

    const msg = data.notification

    const newNotification = {
      key: msg.key,
      type: msg.type.toLowerCase(),
      props: {
        header: msg.header,
        content: msg.content,
        negative: msg.negative,
        positive: msg.positive,
      },
    }

    const notifyIndex = newMessages.findIndex(
      msg => msg.key == newNotification.key
    )
    if (notifyIndex != -1) {
      newMessages[notifyIndex] = newNotification
    } else {
      newMessages.push(newNotification)
    }

    // const update = data.scannerStatusUpdate

    // if (update.success) {
    //   newMessages[0] = {
    //     key: 'primary',
    //     type: 'progress',
    //     props: {
    //       header: update.finished ? 'Synced' : 'Syncing',
    //       content: update.message,
    //       percent: update.progress,
    //       positive: update.finished,
    //     },
    //   }

    //   if (!update.finished) newMessages[0].props.onDismiss = null
    // } else {
    //   const key = Math.random().toString(26)
    //   newMessages.push({
    //     key,
    //     type: 'message',
    //     props: {
    //       header: 'Sync error',
    //       content: update.message,
    //       negative: true,
    //     },
    //   })
    // }

    setMessages(newMessages)
  }, [data, error])

  return null
}

SubscriptionsHook.propTypes = {
  messages: PropTypes.array.isRequired,
  setMessages: PropTypes.func.isRequired,
}

export default SubscriptionsHook
