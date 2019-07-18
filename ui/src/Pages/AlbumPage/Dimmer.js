import React from 'react'
import { useSpring, animated } from 'react-spring'

const Dimmer = ({ onClick, active }) => {
  const [props, set, stop] = useSpring(() => ({ opacity: 0 }))

  set({
    opacity: active ? 1 : 0,
  })

  const AnimatedDimmer = styled(animated.div)`
    position: fixed;
    width: 100%;
    height: 100%;
    background-color: black;
    margin: 0;
    z-index: 10;
  `

  return (
    <AnimatedDimmer
      onClick={onClick}
      style={{
        ...props,
        pointerEvents: active ? 'auto' : 'none',
      }}
    />
  )
}

export default Dimmer
