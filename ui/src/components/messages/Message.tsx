import React from 'react'
import { forwardRef } from 'react'
import { ReactComponent as DismissIcon } from './icons/dismissIcon.svg'

export type MessageProps = {
  header: string
  content?: string
  children?: React.ReactNode
  onDismiss?(): void
}

const Message = forwardRef(
  (
    { onDismiss, header, children, content }: MessageProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    return (
      <div
        ref={ref}
        className="bg-white dark:bg-dark-bg2 shadow-md border rounded p-2 h-[84px] relative"
      >
        <button onClick={onDismiss} className="absolute top-3 right-2">
          <DismissIcon className="w-[10px] h-[10px] text-gray-700 dark:text-gray-200" />
        </button>
        <h1 className="font-semibold text-sm">{header}</h1>
        <div className="text-sm">{content}</div>
        {children}
      </div>
    )
  }
)

export default Message
