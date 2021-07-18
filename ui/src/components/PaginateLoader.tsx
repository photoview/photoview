import React from 'react'
import Loader from '../primitives/Loader'

type PaginateLoaderProps = {
  active: boolean
  text: string
}

const PaginateLoader = ({ active, text }: PaginateLoaderProps) => (
  <Loader
    active
    message={text}
    className={active ? 'opacity-100' : 'opacity-0'}
  />
)

export default PaginateLoader
