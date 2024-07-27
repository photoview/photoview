import React, { createContext, useContext, useState, ReactNode } from 'react'
import { Message } from './SubscriptionsHook'

export type MessageStateType = {
  messages: Message[]
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>
  add(message: Message): void
  removeKey(key: string): void
}

const MessageContext = createContext<MessageStateType | undefined>(undefined)

interface MessageProviderProps {
  children: ReactNode
}

export const MessageProvider: React.FC<MessageProviderProps> = ({ children }) => {
  const [messages, setMessages] = useState<Message[]>([])

  const add = (message: Message) => {
    setMessages(prevMessages => {
      const newMessages = prevMessages.filter(msg => msg.key != message.key)
      newMessages.push(message)
      return newMessages
    })
  }

  const removeKey = (key: string) => {
    setMessages(prevMessages => prevMessages.filter(msg => msg.key != key))
  }

  return (
    <MessageContext.Provider value={{ messages, setMessages, add, removeKey }}>
      {children}
    </MessageContext.Provider>
  )
}

export const useMessageState = (): MessageStateType => {
  const context = useContext(MessageContext)
  if (!context) {
    throw new Error('useMessageState must be used within a MessageProvider')
  }
  return context
}
