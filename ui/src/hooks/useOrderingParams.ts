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

function useOrderingParams(
  { getParam, setParams }: UrlParams,
  defaultOrderBy = 'date_shot'
) {
  const rawOrderBy = getParam('orderBy', defaultOrderBy)
  const orderBy = rawOrderBy === null || rawOrderBy === '' ? defaultOrderBy : rawOrderBy

  const rawOrderDir = getParam('orderDirection', OrderDirection.ASC)
  const orderDirection =
    Object.values(OrderDirection).includes(rawOrderDir as OrderDirection)
      ? (rawOrderDir as OrderDirection) : OrderDirection.ASC

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
