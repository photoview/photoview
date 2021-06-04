import React, { forwardRef } from 'react'
import classNames, { Argument as ClassNamesArg } from 'classnames'

type TextFieldProps = {
  label?: string
  error?: string
  className?: ClassNamesArg
  sizeVariant?: 'default' | 'small'
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, 'className'>

export const TextField = forwardRef(
  (
    { label, error, className, sizeVariant, ...inputProps }: TextFieldProps,
    ref: React.ForwardedRef<HTMLInputElement>
  ) => {
    sizeVariant = sizeVariant ?? 'default'

    let variant = 'bg-white border-gray-200 focus:border-blue-400'
    if (error)
      variant =
        'bg-red-50 border-red-200 focus:border-red-400 focus:ring-red-100'

    const input = (
      <input
        className={classNames(
          'block border rounded-md w-full focus:ring-2 focus:outline-none px-2',
          variant,
          sizeVariant == 'default' ? 'py-2' : 'py-1'
        )}
        {...inputProps}
        ref={ref}
      />
    )

    let errorElm = null
    if (error) errorElm = <div className="text-red-800">{error}</div>

    if (label) {
      return (
        <label className={classNames('block', className)}>
          <span className="block text-xs uppercase font-semibold mb-1">
            {label}
          </span>
          {input}
          {errorElm}
        </label>
      )
    }

    return (
      <div
        className={classNames(className, sizeVariant == 'small' && 'text-sm')}
      >
        {input}
        {errorElm}
      </div>
    )
  }
)

export const Submit = (props: React.InputHTMLAttributes<HTMLInputElement>) => {
  return (
    <input
      className={`rounded-md px-8 py-2 focus:outline-none hover:cursor-pointer bg-[#eee] hover:bg-[#e6e6e6] focus:bg-[#e6e6e6] ${props.className}`}
      type="submit"
      {...props}
    />
  )
}
