import { useCallback, useEffect, useRef } from 'react'

const useScrollPagination = ({ loading, fetchMore }) => {
  const observer = useRef(null)
  const observerElem = useRef(null)

  const reconfigureIntersectionObserver = () => {
    var options = {
      root: null,
      rootMargin: '-100% 0px 0px 0px',
      threshold: 0,
    }

    // delete old observer
    if (observer.current) observer.current.disconnect()

    // configure new observer
    observer.current = new IntersectionObserver(entities => {
      console.log('Observing', entities)
      if (entities.find(x => x.isIntersecting == false)) {
        console.log('load more')
        fetchMore()
      }
    }, options)

    // activate new observer
    if (observerElem.current && !loading)
      observer.current.observe(observerElem.current)
  }

  const containerElem = useCallback(node => {
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
    if (observer.current != null) {
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
  }, [fetchMore])

  return {
    containerElem,
  }
}

export default useScrollPagination
