import React from 'react'
import PropTypes from 'prop-types'
import { Loader } from 'semantic-ui-react'

const PaginateLoader = ({ active, text }) => (
  <Loader
    style={{ margin: '42px 0 24px 0', opacity: active ? '1' : '0' }}
    inline="centered"
    active={true}
  >
    {text}
  </Loader>
)

PaginateLoader.propTypes = {
  active: PropTypes.bool,
  text: PropTypes.string,
}

export default PaginateLoader
