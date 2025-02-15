import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react'
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
    throw new Error('useMessageState was called outside of MessageProvider. Ensure it is used within MessageProvider.')
  }
  return context
}

type MessageProviderProps = {
  children: ReactNode
}

export const MessageProvider = ({ children }: MessageProviderProps) => {
  const [messages, setMessages] = useState<Message[]>([])

  const add = (message: Message) => {
    const timestampedMessage = { ...message, timestamp: Date.now() };
    setMessages((prevMessages) => [...prevMessages, timestampedMessage])
  }

  const removeKey = (key: string) => {
    setMessages((prevMessages) => prevMessages.filter((msg) => msg.key !== key))
  }

  const CLEANUP_INTERVAL = 60 * 60 * 1000; // 1 hour in ms
  const MESSAGE_LIFETIME = 24 * 60 * 60 * 1000; // 24 hours in ms

  useEffect(() => {
    const cleanupInterval = setInterval(() => {
      setMessages((prevMessages) => {
        const cutoff = Date.now() - MESSAGE_LIFETIME

        return prevMessages.filter((msg) => (msg.timestamp ?? 0) > cutoff);
      });
    }, CLEANUP_INTERVAL);

    return () => clearInterval(cleanupInterval);
  }, []);

  return (
    <MessageContext.Provider value={{ messages, setMessages, add, removeKey }}>
      {children}
    </MessageContext.Provider>
  )
}
