import classNames, { Argument as ClassNamesArg } from 'classnames'
import { overrideTailwindClasses } from 'tailwind-override'

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
