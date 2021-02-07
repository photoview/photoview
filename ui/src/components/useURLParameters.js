import { useState } from 'react'

function useURLParameters() {
  const [urlString, setUrlString] = useState(document.location.href)

  const url = new URL(urlString)
  const params = new URLSearchParams(url.search)

  const getParam = (key, defaultValue = null) => {
    return params.has(key) ? params.get(key) : defaultValue
  }

  const updateParams = () => {
    history.replaceState({}, '', url.pathname + '?' + params.toString())
    setUrlString(document.location.href)
  }

  const setParam = (key, value) => {
    params.set(key, value)
    updateParams()
  }

  const setParams = pairs => {
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
