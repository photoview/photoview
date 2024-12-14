import React, { createContext, useContext, useState, ReactNode } from 'react'
import { Message } from './SubscriptionsHook'

type MessageContextType = {
  messages: Message[]
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>
  add: (message: Message) => void
  removeKey: (key: string) => void
}

const MessageContext = createContext<MessageContextType | undefined>(undefined)

export const useMessageState = (): MessageContextType => {
  const context = useContext(MessageContext)
  if (context === undefined) {
    throw new Error('useMessageState must be used within a MessageProvider')
  }
  return context
}

type MessageProviderProps = {
  children: ReactNode
}

export const MessageProvider = ({ children }: MessageProviderProps) => {
  const [messages, setMessages] = useState<Message[]>([])

  const add = (message: Message) => {
    setMessages((prevMessages) => [...prevMessages, message])
  }

  const removeKey = (key: string) => {
    setMessages((prevMessages) => prevMessages.filter((msg) => msg.key !== key))
  }

  return (
    <MessageContext.Provider value={{ messages, setMessages, add, removeKey }}>
      {children}
    </MessageContext.Provider>
  )
}
