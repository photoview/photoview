import React from 'react'
import { Loader } from 'semantic-ui-react'

type PaginateLoaderProps = {
  active: boolean
  text: string
}

const PaginateLoader = ({ active, text }: PaginateLoaderProps) => (
  <Loader
    style={{ margin: '42px 0 24px 0', opacity: active ? '1' : '0' }}
    inline="centered"
    active={true}
  >
    {text}
  </Loader>
)

export default PaginateLoader
