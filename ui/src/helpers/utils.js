export function debounce(func, wait, triggerRising) {
  let timeout = null

  const debounced = (...args) => {
    if (timeout) {
      clearTimeout(timeout)
      timeout = null
    } else if (triggerRising) {
      func(...args)
    }

    timeout = setTimeout(() => {
      timeout = null
      func(...args)
    }, wait)
  }

  debounced.cancel = () => {
    clearTimeout(timeout)
    timeout = null
  }

  return debounced
}

export function isNil(value) {
  return value === undefined || value === null
}
