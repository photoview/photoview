import * as React from 'react'

function SvgExit(props: React.SVGProps<SVGSVGElement>) {
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
        <path d="M2 2l32 32M2 34L34 2" />
      </g>
    </svg>
  )
}

export default SvgExit
