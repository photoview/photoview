import React from 'react'

type MessageBoxProps = {
  message?: string | null
  show?: boolean
  type?: 'neutral' | 'positive' | 'negative'
}

const MessageBox = ({ message, show, type }: MessageBoxProps) => {
  if (!show) return null

  let variant = 'bg-gray-100'
  if (type == 'positive') variant = 'bg-green-200 text-green-900'
  if (type == 'negative') variant = 'bg-red-200 text-red-900'

  return <div className={`py-2 px-3 my-4 rounded-md ${variant}`}>{message}</div>
}

export default MessageBox
