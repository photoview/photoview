import React from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Message, Progress } from 'semantic-ui-react'

const MessageProgress = ({ header, content, percent = 0, ...props }) => {
  const StyledProgress = styled(Progress)`
    position: absolute !important;
    bottom: 0;
    left: 0;
    width: 100%;
  `

  return (
    <Message floating {...props}>
      <Message.Content>
        <Message.Header>{header}</Message.Header>
        {content}
        <StyledProgress
          percent={percent}
          size="tiny"
          attached="bottom"
          indicating
        />
      </Message.Content>
    </Message>
  )
}

MessageProgress.propTypes = {
  header: PropTypes.string,
  content: PropTypes.any,
  percent: PropTypes.number,
}

export default MessageProgress
