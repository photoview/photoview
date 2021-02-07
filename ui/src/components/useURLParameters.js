import { useState } from 'react'

function useURLParameters() {
  const [urlString, setUrlString] = useState(document.location.href)

  const url = new URL(urlString)
  const params = new URLSearchParams(url.search)

  const getParam = key => {
    return params.get(key)
  }

  const setParam = (key, value) => {
    params.set(key, value)
    history.replaceState({}, '', url.pathname + '?' + params.toString())

    setUrlString(document.location.href)
  }

  return {
    getParam,
    setParam,
  }
}

export default useURLParameters
