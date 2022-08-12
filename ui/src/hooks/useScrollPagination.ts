import { ApolloQueryResult } from '@apollo/client'
import { useCallback, useEffect, useRef, useState } from 'react'

interface ScrollPaginationArgs<D> {
  loading: boolean
  data: D | undefined
  fetchMore: (args: {
    variables: { offset: number }
  }) => Promise<ApolloQueryResult<D>>
  getItems: (data: D) => unknown[]
}

type ScrollPaginationResult = {
  finished: boolean
  containerElem: (node: null | Element) => void
}

const useScrollPagination: <D>(
  args: ScrollPaginationArgs<D>
) => ScrollPaginationResult = ({ loading, fetchMore, data, getItems }) => {
  const observer = useRef<IntersectionObserver | null>(null)
  const observerElem = useRef<Element | null>(null)
  const [finished, setFinished] = useState(false)

  const reconfigureIntersectionObserver = () => {
    const options = {
      root: null,
      rootMargin: '-100% 0px 0px 0px',
      threshold: 0,
    }

    // delete old observer
    observer.current?.disconnect()

    if (finished) return

    // configure new observer
    observer.current = new IntersectionObserver(entities => {
      if (entities.find(x => x.isIntersecting == false)) {
        const itemCount = data !== undefined ? getItems(data).length : 0
        fetchMore({
          variables: {
            offset: itemCount,
          },
        }).then(result => {
          const newItemCount = getItems(result.data).length
          if (newItemCount == 0) {
            setFinished(true)
          }
        })
      }
    }, options)

    // activate new observer
    if (observerElem.current && !loading) {
      observer.current.observe(observerElem.current)
    }
  }

  const containerElem = useCallback((node: null | Element): void => {
    observerElem.current = node

    // cleanup
    if (observer.current != null) {
      observer.current.disconnect()
    }

    if (node != null) {
      reconfigureIntersectionObserver()
    }
  }, [])

  // only observe when not loading
  useEffect(() => {
    if (observer.current && observerElem.current) {
      if (loading) {
        observer.current.unobserve(observerElem.current)
      } else {
        observer.current.observe(observerElem.current)
      }
    }
  }, [loading])

  // reconfigure observer if fetchMore function changes
  useEffect(() => {
    reconfigureIntersectionObserver()
  }, [fetchMore, data, finished])

  useEffect(() => {
    setFinished(false)
  }, [data])

  return {
    containerElem,
    finished,
  }
}

export default useScrollPagination
