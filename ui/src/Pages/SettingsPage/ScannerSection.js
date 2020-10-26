import React from 'react'
import { useMutation, gql } from '@apollo/client'
import { Button, Icon } from 'semantic-ui-react'
import PeriodicScanner from './PeriodicScanner'
import ScannerConcurrentWorkers from './ScannerConcurrentWorkers'
import { SectionTitle, InputLabelDescription } from './SettingsPage'

const SCAN_MUTATION = gql`
  mutation scanAllMutation {
    scanAll {
      success
      message
    }
  }
`

const ScannerSection = () => {
  const [startScanner, { called }] = useMutation(SCAN_MUTATION)

  return (
    <div>
      <SectionTitle nospace>Scanner</SectionTitle>
      <InputLabelDescription>
        Will scan all users for new or updated media
      </InputLabelDescription>
      <Button
        icon
        labelPosition="left"
        onClick={() => {
          startScanner()
        }}
        disabled={called}
      >
        <Icon name="sync" />
        Scan all users
      </Button>
      <PeriodicScanner />
      <ScannerConcurrentWorkers />
    </div>
  )
}

export default ScannerSection
