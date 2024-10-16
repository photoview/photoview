import React from 'react'
import { forwardRef } from 'react'
import { ReactComponent as DismissIcon } from './icons/dismissIcon.svg'

export type MessageProps = {
  header: string
  content?: string
  children?: React.ReactNode
  onDismiss?(): void
  negative?: boolean
  positive?: boolean
}

const Message = forwardRef(
  (
    { onDismiss, header, children, content, negative, positive }: MessageProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    let backgroundColor = 'bg-white dark:bg-dark-bg2'
    if (negative) backgroundColor = 'bg-red-100 dark:bg-red-900'
    if (positive) backgroundColor = 'bg-green-100 dark:bg-green-900'

    return (
      <div
        ref={ref}
        className={`${backgroundColor} shadow-md border rounded p-2 relative`}
      >
        <button onClick={onDismiss} className="absolute top-3 right-2">
          <DismissIcon className="w-[10px] h-[10px] text-gray-700 dark:text-gray-200" />
        </button>
        <h1 className="font-semibold text-sm">{header}</h1>
        <div
          className="text-sm overflow-y-auto"
          style={{
            maxHeight: '6rem', // approx. 5 lines
            overflowY: 'auto',
            whiteSpace: 'pre-wrap',
            wordWrap: 'break-word',
            overflowWrap: 'break-word'
          }}
        >
          {content}
        </div>
        {children}
      </div>
    )
  }
)

export default Message
