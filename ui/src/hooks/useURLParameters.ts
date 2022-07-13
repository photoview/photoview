import { useState } from 'react'

export type UrlKeyValuePair = { key: string; value: string | null }

export type UrlParams = {
  getParam: (key: string, defaultValue?: string | null) => string | null
  setParam: (key: string, value: string | null) => void
  setParams: (pairs: UrlKeyValuePair[]) => void
}

const useURLParameters: () => UrlParams = () => {
  const [urlString, setUrlString] = useState(document.location.href)

  const url = new URL(urlString)
  const params = new URLSearchParams(url.search)

  const getParam = (key: string, defaultValue: string | null = null) => {
    return params.has(key) ? params.get(key) : defaultValue
  }

  const updateParams = () => {
    if (params.toString()) {
      history.replaceState({}, '', url.pathname + '?' + params.toString())
    } else {
      history.replaceState({}, '', url.pathname)
    }
    setUrlString(document.location.href)
  }

  const setParam = (key: string, value: string | null) => {
    if (value) {
      params.set(key, value)
    } else {
      params.delete(key)
    }
    updateParams()
  }

  const setParams = (pairs: UrlKeyValuePair[]) => {
    for (const pair of pairs) {
      if (pair.value) {
        params.set(pair.key, pair.value)
      } else {
        params.delete(pair.key)
      }
    }
    updateParams()
  }

  return {
    getParam,
    setParam,
    setParams,
  }
}

export default useURLParameters
