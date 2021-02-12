import { useCallback, useEffect, useRef, useState } from 'react'

const useScrollPagination = ({ loading, fetchMore, data, getItems }) => {
  const observer = useRef(null)
  const observerElem = useRef(null)
  const [finished, setFinished] = useState(false)

  const reconfigureIntersectionObserver = () => {
    var options = {
      root: null,
      rootMargin: '-100% 0px 0px 0px',
      threshold: 0,
    }

    // delete old observer
    if (observer.current) observer.current.disconnect()

    if (finished) return

    // configure new observer
    observer.current = new IntersectionObserver(entities => {
      console.log('Observing', entities)
      if (entities.find(x => x.isIntersecting == false)) {
        let itemCount = getItems(data).length
        console.log('load more', itemCount)
        fetchMore({
          variables: {
            offset: itemCount,
          },
        }).then(result => {
          const newItemCount = getItems(result.data).length
          console.log('then', result, itemCount, newItemCount)
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
  }, [fetchMore, data, finished])

  return {
    containerElem,
    finished,
  }
}

export default useScrollPagination
