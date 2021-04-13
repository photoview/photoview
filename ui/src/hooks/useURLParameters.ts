import { useState } from 'react'

export type UrlKeyValuePair = { key: string; value: string }

function useURLParameters() {
  const [urlString, setUrlString] = useState(document.location.href)

  const url = new URL(urlString)
  const params = new URLSearchParams(url.search)

  const getParam = (key: string, defaultValue: string | null = null) => {
    return params.has(key) ? params.get(key) : defaultValue
  }

  const updateParams = () => {
    history.replaceState({}, '', url.pathname + '?' + params.toString())
    setUrlString(document.location.href)
  }

  const setParam = (key: string, value: string) => {
    params.set(key, value)
    updateParams()
  }

  const setParams = (pairs: UrlKeyValuePair[]) => {
    for (const pair of pairs) {
      params.set(pair.key, pair.value)
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
