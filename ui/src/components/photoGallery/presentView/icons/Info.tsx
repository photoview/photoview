import * as React from 'react'

function SvgInfo(props: React.SVGProps<SVGSVGElement>) {
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
        <path d="M3,18a15,15 0 1,0 30,0a15,15 0 1,0 -30,0" />
        <path d="M18 16L18 27" />
        <path d="M18 9L18 9.5" />
      </g>
    </svg>
  )
}

export default SvgInfo
