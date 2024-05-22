import * as React from 'react'

function SvgPhoto(props: React.SVGProps<SVGSVGElement>) {
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
        <path d="M 84.66,19.062 H 78.661 V 78.725 H 18.998 v 5.999 H 84.66 Z m 4.276,10.276 V 89.001 H 29.274 V 95 H 94.936 V 29.338 Z M 5,73.662 H 73.662 V 5 H 5 Z M 21.592,56.3 c 1.813,-8.195 9.131,-14.344 17.863,-14.344 8.732,0 16.05,6.149 17.863,14.344 z M 32.571,30.272 c 0,-3.795 3.089,-6.883 6.884,-6.883 3.795,0 6.883,3.088 6.883,6.883 0,3.796 -3.088,6.884 -6.883,6.884 -3.795,0 -6.884,-3.088 -6.884,-6.884 z M 67.662,56.3 H 62.207 C 60.798,48.172 55.131,41.488 47.591,38.64 c 2.184,-2.124 3.547,-5.088 3.547,-8.368 0,-6.442 -5.241,-11.683 -11.683,-11.683 -6.442,0 -11.683,5.241 -11.683,11.683 0,3.28 1.363,6.244 3.547,8.368 -7.54,2.848 -13.208,9.532 -14.616,17.66 H 10.999 V 10.999 h 56.663 z" />
      </g>
    </svg>
  )
}

export default SvgPhoto
