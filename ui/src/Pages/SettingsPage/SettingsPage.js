import React from 'react'
import Layout from '../../Layout'

import { Button, Icon } from 'semantic-ui-react'
import { Mutation } from 'react-apollo'
import gql from 'graphql-tag'
import UsersTable from './UsersTable'

const scanMutation = gql`
  mutation scanAllMutation {
    scanAll {
      success
      message
    }
  }
`

const SettingsPage = () => (
  <Layout>
    <h1>Settings</h1>
    <Mutation mutation={scanMutation}>
      {(scan, { data, called }) => (
        <>
          <h2>Scanner</h2>
          <Button
            icon
            labelPosition="left"
            onClick={() => {
              scan()
            }}
            disabled={called}
          >
            <Icon name="sync" />
            Scan All
          </Button>
        </>
      )}
    </Mutation>
    <UsersTable />
  </Layout>
)

export default SettingsPage
