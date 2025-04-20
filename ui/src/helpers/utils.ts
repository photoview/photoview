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

/**
 * Determines whether a value is `undefined` or `null`.
 *
 * @param value - The value to check.
 * @returns `true` if the value is `undefined` or `null`; otherwise, `false`.
 */
export function isNil(value: unknown): value is undefined | null {
  return value === undefined || value === null
}

/**
 * Throws an error if called, indicating that an unexpected value was encountered in an exhaustive type check.
 *
 * @param value - The unexpected value that triggered the exhaustive check failure.
 * @throws {Error} Always throws to signal that a value of type {@link never} was unexpectedly received.
 */
export function exhaustiveCheck(value: never): never {
  // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
  throw new Error(`Exhaustive check failed with value: ${value}`)
}

export function tailwindClassNames(...args: ClassNamesArg[]) {
  return overrideTailwindClasses(classNames(args))
  // return classNames(args)
}
