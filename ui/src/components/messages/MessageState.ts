import React from 'react'
import { Message } from './SubscriptionsHook'

export type MessageStateType = {
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
      const newMessages = messages.filter(msg => msg.key !== message.key)
      newMessages.push(message)

      return newMessages
    })
  },
  removeKey: key => {
    MessageState.set(messages => {
      const newMessages = messages.filter(msg => msg.key !== key)
      return newMessages
    })
  },
}
