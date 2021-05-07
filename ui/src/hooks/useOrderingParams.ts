import { useCallback } from 'react'
import { OrderDirection } from '../../__generated__/globalTypes'
import { UrlKeyValuePair, UrlParams } from './useURLParameters'

function useOrderingParams({ getParam, setParams }: UrlParams) {
  const orderBy = getParam('orderBy', 'date_shot')

  const orderDirStr = getParam('orderDirection', 'ASC') || 'hello'
  const orderDirection = orderDirStr as OrderDirection

  type setOrderingFn = (args: {
    orderBy?: string
    orderDirection?: OrderDirection
  }) => void

  const setOrdering: setOrderingFn = useCallback(
    ({ orderBy, orderDirection }) => {
      const updatedParams: UrlKeyValuePair[] = []
      if (orderBy !== undefined) {
        updatedParams.push({ key: 'orderBy', value: orderBy })
      }
      if (orderDirection !== undefined) {
        updatedParams.push({ key: 'orderDirection', value: orderDirection })
      }

      setParams(updatedParams)
    },
    [setParams]
  )

  return {
    orderBy,
    orderDirection,
    setOrdering,
  }
}

export default useOrderingParams
