import React, { forwardRef } from 'react'
import MessagePlain, { MessageProps } from './Message'

type MessageProgressProps = MessageProps & {
  percent?: number
}

const MessageProgress = forwardRef(
  (
    { header, content, percent = 0, ...props }: MessageProgressProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    let color = '#dc2625'
    if (percent > 33) color = '#fbbf24'
    if (percent > 66) color = '#56e263'

    return (
      <MessagePlain header={header} content={content} {...props} ref={ref}>
        <div className="absolute bottom-0 left-0 right-0 h-[3px] rounded-b overflow-hidden">
          <div
            className="h-full transition-all duration-200"
            style={{ width: `${percent}%`, backgroundColor: color }}
          ></div>
        </div>
      </MessagePlain>
    )
  }
)

export default MessageProgress
