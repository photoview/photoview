import React from 'react'

type TextFieldProps = {
  label?: string
  error?: string
} & React.InputHTMLAttributes<HTMLInputElement>

export const TextField = ({
  label,
  error,
  className,
  ...inputProps
}: TextFieldProps) => {
  const input = (
    <input
      className="block bg-white border border-gray-200 rounded-md h-10 w-full focus:border-blue-400 focus:ring-2 focus:outline-none px-2"
      {...inputProps}
    />
  )

  let errorElm = null
  if (error) errorElm = <div>{error}</div>

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

export const Submit = (props: React.InputHTMLAttributes<HTMLInputElement>) => {
  return (
    <input
      className={`rounded-md px-8 py-2 focus:outline-none hover:cursor-pointer bg-[#eee] hover:bg-[#e6e6e6] focus:bg-[#e6e6e6] ${props.className}`}
      type="submit"
      {...props}
    />
  )
}
