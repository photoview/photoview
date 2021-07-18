import { useCallback } from 'react'
import { OrderDirection } from '../__generated__/globalTypes'
import { UrlKeyValuePair, UrlParams } from './useURLParameters'

export type MediaOrdering = {
  orderBy: string | null
  orderDirection: OrderDirection | null
}

export type SetOrderingFn = (args: {
  orderBy?: string
  orderDirection?: OrderDirection
}) => void

function useOrderingParams({ getParam, setParams }: UrlParams) {
  const orderBy = getParam('orderBy', 'date_shot')

  const orderDirStr = getParam('orderDirection', 'ASC') || 'hello'
  const orderDirection = orderDirStr as OrderDirection

  const setOrdering: SetOrderingFn = useCallback(
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
