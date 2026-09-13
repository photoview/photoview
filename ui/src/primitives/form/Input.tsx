import React, { forwardRef, useEffect, useState } from 'react'
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

    // A field that stops being a password one and later becomes one again
    // would otherwise come back revealed, showing the secret unasked.
    useEffect(() => {
      if (!isPassword) {
        setRevealed(false)
      }
    }, [isPassword])

    let variant = 'bg-white border-gray-200 focus:border-blue-400'
    if (error)
      variant =
        'bg-red-50 border-red-200 focus:border-red-400 focus:ring-red-100 placeholder-red-300'

    if (disabled) variant = 'bg-gray-100'

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
            'pr-8': (action || isPassword) && !(action && isPassword),
            'pr-16': action && isPassword,
          },
          className
        )}
        {...inputProps}
        type={effectiveType}
        ref={ref}
      />
    )

    if (loading) {
      input = (
        <div className="relative">
          {input}
          <LoadingSpinnerIcon
            aria-label="Loading"
            className="absolute right-[8px] top-[7px] animate-spin"
          />
        </div>
      )
    } else if (action || isPassword) {
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
                disabled={disabled}
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
            {action && (
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
