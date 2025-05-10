import * as React from 'react'

function SvgVideo(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="25 25 36 36"
      fillRule="evenodd"
      clipRule="evenodd"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeMiterlimit={1.5}
      width="1em"
      height="1em"
      transform="scale(.4)"
      {...props}
    >
      <g fill="none" stroke="#000" strokeWidth={3}>
      <path
        d="m -10.5,10.5 v 69.2 h 95 v -69.2 z m 43.2,4.2 h 8.7 v 7.1 h -8.7 z m -13,0 h 8.7 v 7.1 h -8.7 z m -17.2,60.8 h -8.7 v -7.1 h 8.7 z m 0,-53.6 h -8.7 v -7.1 h 8.7 z m 13,53.6 h -8.7 v -7.1 h 8.7 z m 0,-53.6 h -8.7 v -7.1 h 8.7 z m 13,53.6 h -8.7 v -7.1 h 8.7 z m -2,-18 v -24.8 c 0,-2.1 2.3,-3.4 4.1,-2.3 l 19.8,12.4 c 1.7,1.1 1.7,3.5 0,4.6 l -19.8,12.4 c -1.8,1.1 -4.1,-0.2 -4.1,-2.3 z m 15,18 h -8.7 v -7.1 h 8.7 z m 13,0 h -8.7 v -7.1 h 8.7 z m 0,-53.6 h -8.7 v -7.1 h 8.7 z m 12.9,53.6 h -8.7 v -7.1 h 8.7 z m 0,-53.6 h -8.7 v -7.1 h 8.7 z m 12.8,53.6 h -8.7 v -7.1 h 8.7 z m 0,-53.6 h -8.7 v -7.1 h 8.7 z"
      />
      </g>

    </svg>
  )
}

export default SvgVideo
