import React, { forwardRef } from 'react'

type TextFieldProps = {
  label?: string
  error?: string
} & React.InputHTMLAttributes<HTMLInputElement>

export const TextField = forwardRef(
  (
    { label, error, className, ...inputProps }: TextFieldProps,
    ref: React.ForwardedRef<HTMLInputElement>
  ) => {
    let variant = 'bg-white border-gray-200 focus:border-blue-400'
    if (error)
      variant =
        'bg-red-50 border-red-200 focus:border-red-400 focus:ring-red-100'

    const input = (
      <input
        className={`block border rounded-md h-10 w-full  focus:ring-2 focus:outline-none px-2 ${variant}`}
        {...inputProps}
        ref={ref}
      />
    )

    let errorElm = null
    if (error) errorElm = <div className="text-red-800">{error}</div>

    if (label) {
      return (
        <label className={`block my-4 ${className}`}>
          <span className="block text-xs uppercase font-semibold mb-1">
            {label}
          </span>
          {input}
          {errorElm}
        </label>
      )
    }

    return (
      <div className={className}>
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
