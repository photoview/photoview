import * as React from 'react'

function SvgNext(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 28 52"
      fillRule="evenodd"
      clipRule="evenodd"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeMiterlimit={1.5}
      width="1em"
      height="1em"
      {...props}
    >
      <path d="M2 2l24 24L2 50" fill="none" stroke="#000" strokeWidth={3} />
    </svg>
  )
}

export default SvgNext
