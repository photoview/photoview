import React, { CSSProperties } from 'react'

type LoaderProps = {
  active: boolean
  message?: string
  size?: 'small' | 'default'
  className?: string
  style?: CSSProperties
}

const Loader = ({ active, message, className, style }: LoaderProps) => {
  if (!active) return null
  return (
    <div className={className} style={style}>
      {message ?? 'Loading...'}
    </div>
  )
}

export default Loader
