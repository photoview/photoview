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
    const PROGRESS_LEVELS = {
      LOW: { threshold: 0, color: '#dc2625', state: 'low progress' },
      MEDIUM: { threshold: 33, color: '#fbbf24', state: 'medium progress' },
      HIGH: { threshold: 66, color: '#56e263', state: 'high progress' }
    } as const;

    type ProgressLevel = typeof PROGRESS_LEVELS[keyof typeof PROGRESS_LEVELS];
    type ProgressColor = ProgressLevel['color'];
    type ProgressState = ProgressLevel['state'];

    let color: ProgressColor = PROGRESS_LEVELS.LOW.color
    let state: ProgressState = PROGRESS_LEVELS.LOW.state
    if (percent > PROGRESS_LEVELS.MEDIUM.threshold) {
      color = PROGRESS_LEVELS.MEDIUM.color
      state = PROGRESS_LEVELS.MEDIUM.state
    }
    if (percent > PROGRESS_LEVELS.HIGH.threshold) {
      color = PROGRESS_LEVELS.HIGH.color
      state = PROGRESS_LEVELS.HIGH.state
    }

    return (
      <MessagePlain header={header} content={content} {...props} ref={ref}>
        <div className="absolute bottom-0 left-0 right-0 h-[3px] rounded-b overflow-hidden">
          <div
            role="progressbar"
            aria-valuenow={percent}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`${state}: ${percent}%`}
            className="h-full transition-all duration-200"
            style={{ width: `${percent}%`, backgroundColor: color }}
          />
        </div>
      </MessagePlain>
    )
  }
)

export default MessageProgress
