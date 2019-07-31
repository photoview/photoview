import React from 'react'
import Layout from '../../Layout'

import { Button, Icon } from 'semantic-ui-react'
import { Mutation } from 'react-apollo'
import gql from 'graphql-tag'

const scanMutation = gql`
  mutation scanAllMutation {
    scanAll {
      success
      errorMessage
    }
  }
`

const SettingsPage = () => (
  <Layout>
    <h1>Settings</h1>
    <Mutation mutation={scanMutation}>
      {(scan, { data }) => (
        <>
          <h2>Scanner</h2>
          <Button
            icon
            labelPosition="left"
            onClick={() => {
              scan()
            }}
            disabled={data && data.scanAll && data.scanAll.success}
          >
            <Icon name="sync" />
            Scan All
          </Button>
          <p>Scan for new images for all users</p>
        </>
      )}
    </Mutation>
  </Layout>
)

export default SettingsPage
