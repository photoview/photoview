import { useLayoutEffect, useState, useRef } from 'react'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function useDelay(wait: number, deps: any[] = []) {
  const triggerUpdate = useState(false)[1]
  const done = useRef(false)

  useLayoutEffect(() => {
    const handle = setTimeout(() => {
      done.current = true
      triggerUpdate(x => !x)
    }, wait)

    return () => {
      done.current = false
      clearTimeout(handle)
    }
  }, deps)

  return done.current
}

export default useDelay
