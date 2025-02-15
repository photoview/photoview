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
    const backgroundColorClass = negative
      ? 'bg-red-100 dark:bg-red-900'
      : positive
        ? 'bg-green-100 dark:bg-green-900'
        : 'bg-white dark:bg-dark-bg2'

    return (
      <div
        ref={ref}
        className={`${backgroundColorClass} shadow-md border rounded p-2 relative`}
      >
        <button type="button" onClick={onDismiss} className="absolute top-3 right-2">
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
