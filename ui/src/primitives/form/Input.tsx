import React, { forwardRef, useEffect, useRef, useState } from 'react'
import classNames, { Argument as ClassNamesArg } from 'classnames'
import { ReactComponent as ActionArrowIcon } from './icons/textboxActionArrow.svg'
import { ReactComponent as LoadingSpinnerIcon } from './icons/textboxLoadingSpinner.svg'
import { ReactComponent as EyeIcon } from './icons/eyeIcon.svg'
import { ReactComponent as EyeOffIcon } from './icons/eyeOffIcon.svg'
import styled from 'styled-components'
import { tailwindClassNames } from '../../helpers/utils'

type TextFieldProps = {
  label?: string
  error?: string
  className?: ClassNamesArg
  wrapperClassName?: ClassNamesArg
  sizeVariant?: 'default' | 'big'
  fullWidth?: boolean
  action?: () => void
  loading?: boolean
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, 'className'>

export const TextField = forwardRef(
  (
    {
      label,
      error,
      className,
      wrapperClassName,
      sizeVariant,
      fullWidth,
      action,
      loading,
      type,
      ...inputProps
    }: TextFieldProps,
    ref: React.ForwardedRef<HTMLInputElement>
  ) => {
    const disabled = !!inputProps.disabled
    sizeVariant = sizeVariant ?? 'default'

    const isPassword = type === 'password'
    const [revealed, setRevealed] = useState(false)
    const effectiveType = isPassword && revealed ? 'text' : type

    // `revealed` only has an effect while the field is a password field: it is
    // what turns the password's type into text. Once the field stops being a
    // password field it has nothing to reveal, so the flag is cleared then -
    // otherwise a field that later becomes a password field again would
    // arrive already revealed, showing the secret without anyone asking.
    useEffect(() => {
      if (!isPassword) {
        setRevealed(false)
      }
    }, [isPassword])

    // The input's own ref, alongside the one a caller may have forwarded, so
    // the submit handler below can reach the element and its form.
    const inputRef = useRef<HTMLInputElement | null>(null)
    const setInputRef = (element: HTMLInputElement | null) => {
      inputRef.current = element
      if (typeof ref === 'function') {
        ref(element)
      } else if (ref) {
        ref.current = element
      }
    }

    // A revealed password is a text input, and browsers may remember what was
    // typed into a text input as an autocomplete suggestion once its form is
    // submitted. Masking it again on submit keeps it out of that history. The
    // type is set on the element directly, inside the event, because
    // re-rendering from state would only happen after the browser has already
    // looked at the field.
    useEffect(() => {
      const form = inputRef.current?.form
      if (!isPassword || !revealed || !form) return

      const mask = () => {
        if (inputRef.current) inputRef.current.type = 'password'
        setRevealed(false)
      }

      form.addEventListener('submit', mask)

      return () => form.removeEventListener('submit', mask)
    }, [isPassword, revealed])

    let variant = 'bg-white border-gray-200 focus:border-blue-400'
    if (error)
      variant =
        'bg-red-50 border-red-200 focus:border-red-400 focus:ring-red-100 placeholder-red-300'

    if (disabled) variant = 'bg-gray-100'

    // What sits inside the field on the right: the reveal button, and either
    // the action button or the spinner that stands in for it while loading.
    // The field's right padding has to leave room for however many there are.
    const trailingControls = (isPassword ? 1 : 0) + (loading || action ? 1 : 0)

    let keyUpEvent = undefined
    if (action) {
      keyUpEvent = (event: React.KeyboardEvent<HTMLInputElement>) => {
        if (inputProps.onKeyUp) inputProps.onKeyUp(event)

        if (event.key == 'Enter') {
          event.preventDefault()
          action()
        }
      }
    }

    let input = (
      <input
        onKeyUp={keyUpEvent}
        className={classNames(
          'block border rounded-md focus:ring-2 focus:outline-none px-2',
          'dark:bg-dark-input-bg dark:border-dark-input-border',
          variant,
          sizeVariant == 'big' ? 'py-2' : 'py-1',
          {
            'w-full': fullWidth,
            'pr-8': trailingControls == 1,
            'pr-16': trailingControls > 1,
          },
          className
        )}
        {...inputProps}
        type={effectiveType}
        // A revealed password is plain text as far as the browser is
        // concerned, and enhanced spellcheck in Chrome and Edge sends the
        // contents of text fields to Google and Microsoft. These are applied
        // after the caller's props so they cannot be switched back on.
        {...(isPassword && {
          spellCheck: false,
          autoCorrect: 'off',
          autoCapitalize: 'off',
        })}
        ref={setInputRef}
      />
    )

    if (loading || action || isPassword) {
      const iconClassName = classNames(
        sizeVariant == 'big' && 'w-4 h-4 mt-1 mr-1'
      )

      input = (
        <div
          className={classNames('relative inline-block', {
            'w-full': fullWidth,
          })}
        >
          {input}
          <div className="absolute top-1/2 right-0 -translate-y-1/2 flex items-center">
            {isPassword && (
              <button
                type="button"
                // A disabled field cannot be revealed, but one that already
                // is must still be hideable - PasswordProtectedShare disables
                // the field while its request runs.
                disabled={disabled && !revealed}
                aria-label={revealed ? 'Hide password' : 'Show password'}
                aria-pressed={revealed}
                className="p-2 text-gray-600 disabled:text-gray-400 disabled:cursor-default"
                onMouseDown={e => e.preventDefault()}
                onClick={() => setRevealed(r => !r)}
              >
                {revealed ? (
                  <EyeOffIcon className={iconClassName} />
                ) : (
                  <EyeIcon className={iconClassName} />
                )}
              </button>
            )}
            {loading ? (
              <LoadingSpinnerIcon
                aria-label="Loading"
                className={classNames('mr-2 animate-spin', iconClassName)}
              />
            ) : (
              action && (
                <button
                  disabled={disabled}
                  aria-label="Submit"
                  className="p-2 text-gray-600 disabled:text-gray-400 disabled:cursor-default"
                  onClick={e => {
                    e.preventDefault()
                    e.stopPropagation()
                    action()
                    return false
                  }}
                >
                  <ActionArrowIcon className={iconClassName} />
                </button>
              )
            )}
          </div>
        </div>
      )
    }

    let errorElm = null
    if (error) errorElm = <div className="text-red-800">{error}</div>

    const wrapperClasses = classNames(
      sizeVariant == 'default' && 'text-sm',
      wrapperClassName
    )

    if (label) {
      return (
        <label className={classNames(wrapperClasses, 'block')}>
          <span className="block text-xs uppercase font-semibold mb-1">
            {label}
          </span>
          {input}
          {errorElm}
        </label>
      )
    }

    return (
      <div className={wrapperClasses}>
        {input}
        {errorElm}
      </div>
    )
  }
)

type ButtonProps = {
  variant?: 'negative' | 'positive' | 'default'
  background?: 'default' | 'white'
  className?: string
}

export const buttonStyles = ({ variant, background }: ButtonProps) =>
  classNames(
    'px-6 py-0.5 rounded border border-gray-200 focus:outline-none focus:border-blue-300 text-[#222] hover:bg-gray-100 whitespace-nowrap',
    'dark:bg-dark-input-bg dark:border-dark-input-border dark:text-dark-input-text dark:focus:border-blue-300',
    variant == 'negative' &&
      'text-red-600 hover:bg-red-600 hover:border-red-700 hover:text-white transition-colors focus:border-red-600 focus:hover:border-red-700',
    variant == 'positive' &&
      'text-green-600 hover:bg-green-600 hover:border-green-700 hover:text-white transition-colors focus:border-green-600 focus:hover:border-green-700',
    background == 'white' ? 'bg-white' : 'bg-gray-50'
  )

type SubmitProps = ButtonProps & {
  children: string
  className?: string
}

export const Submit = ({
  className,
  variant,
  background,
  children,
  ...props
}: SubmitProps & React.ButtonHTMLAttributes<HTMLInputElement>) => (
  <input
    className={tailwindClassNames(
      buttonStyles({ variant, background }),
      className
    )}
    type="submit"
    value={children}
    {...props}
  />
)

export const Button = ({
  children,
  variant,
  background,
  className,
  ...props
}: ButtonProps & React.ButtonHTMLAttributes<HTMLButtonElement>) => (
  <button
    className={tailwindClassNames(
      buttonStyles({ variant, background }),
      className
    )}
    {...props}
  >
    {children}
  </button>
)

export const ButtonGroup = styled.div.attrs({ className: 'flex gap-1' })``
