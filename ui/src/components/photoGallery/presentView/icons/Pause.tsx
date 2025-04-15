import * as React from 'react'

function SvgPause(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 36 36"
      fillRule="evenodd"
      clipRule="evenodd"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeMiterlimit={1.5}
      width="1em"
      height="1em"
      {...props}
    >
      <g fill="none" stroke="#000" strokeWidth={3}>
        <path d="M8 2l0 32 M28 2L28 34" />
      </g>
    </svg>
  )
}

export default SvgPause
