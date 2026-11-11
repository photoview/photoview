import classNames, { Argument as ClassNamesArg } from 'classnames'
import { overrideTailwindClasses } from 'tailwind-override'
import { authToken } from './authentication'

export interface DebouncedFn<F extends (...args: unknown[]) => unknown> {
  (...args: Parameters<F>): void
  cancel(): void
}

export function debounce<F extends (...args: unknown[]) => unknown>(
  func: F,
  wait: number,
  triggerRising?: boolean
): DebouncedFn<F> {
  let timeout: number | undefined = undefined

  const debounced = (...args: Parameters<F>) => {
    if (timeout) {
      clearTimeout(timeout)
      timeout = undefined
    } else if (triggerRising) {
      func(...args)
    }

    timeout = window.setTimeout(() => {
      timeout = undefined
      func(...args)
    }, wait)
  }

  debounced.cancel = () => {
    clearTimeout(timeout)
    timeout = undefined
  }

  return debounced
}

export function isNil(value: unknown): value is undefined | null {
  return value === undefined || value === null
}

export function exhaustiveCheck(value: never) {
  // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
  throw new Error(`Exhaustive check failed with value: ${value}`)
}

export function tailwindClassNames(...args: ClassNamesArg[]) {
  return overrideTailwindClasses(classNames(args))
  // return classNames(args)
}

export function getPublicUrl(url: string = '') {
  try {
    try {
      return new URL(url);
    } catch {
      return new URL(url, import.meta.env.BASE_URL);
    } 
  } catch {
      return new URL(`${import.meta.env.BASE_URL}${url}`.replace(/\/\//g, '/'), location.origin);
  }
}

export function getProtectedUrl<S extends string | undefined>(url: S) {
  if (url == undefined) return undefined as S

  const publicUrl = getPublicUrl(url);

  if (authToken() == null) {
    const tokenRegex = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
    if (tokenRegex) {
      publicUrl.searchParams.set('token', tokenRegex[1])
    }
  }

  return publicUrl.href
}
