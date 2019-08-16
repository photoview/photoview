import React, { Component } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { Message, Progress } from 'semantic-ui-react'
import gql from 'graphql-tag'
import { Subscription } from 'react-apollo'

const syncSubscription = gql`
  subscription syncSubscription {
    scannerStatusUpdate {
      finished
      success
      errorMessage
      progress
    }
  }
`

const Container = styled.div`
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 500px;
`

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
  content: PropTypes.element,
  percent: PropTypes.number,
}

class Messages extends Component {
  constructor(props) {
    super(props)

    this.state = {
      showSyncMessage: true,
    }
  }

  render() {
    if (!localStorage.getItem('token')) {
      return null
    }

    return (
      <Container>
        <Subscription
          subscription={syncSubscription}
          shouldResubscribe
          onSubscriptionData={() => {
            this.setState({ showSyncMessage: true })
          }}
        >
          {({ loading, error, data }) => {
            if (error) return <div>error {error.message}</div>
            if (loading) return null

            console.log('Data update', data)
            const update = data.scannerStatusUpdate

            return (
              <MessageProgress
                hidden={!this.state.showSyncMessage}
                onDismiss={() => {
                  this.setState({ showSyncMessage: false })
                }}
                header={update.finished ? 'Synced' : 'Syncing'}
                content={
                  update.finished ? 'Finished syncing' : 'Syncing in progress'
                }
                percent={update.progress}
              />
            )
          }}
        </Subscription>
      </Container>
    )
  }
}

export default Messages
