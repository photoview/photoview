import React from 'react'

type LoaderProps = {
  active: boolean
  message?: string
}

const Loader = ({ active, message }: LoaderProps) => {
  if (!active) return null
  return <div>{message ?? 'Loading...'}</div>
}

export default Loader
