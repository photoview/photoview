import React from 'react'
import { useMutation, gql } from '@apollo/client'
import { Button, Icon } from 'semantic-ui-react'
import PeriodicScanner from './PeriodicScanner'
import ScannerConcurrentWorkers from './ScannerConcurrentWorkers'
import { SectionTitle, InputLabelDescription } from './SettingsPage'
import { useTranslation } from 'react-i18next'

const SCAN_MUTATION = gql`
  mutation scanAllMutation {
    scanAll {
      success
      message
    }
  }
`

const ScannerSection = () => {
  const { t } = useTranslation()
  const [startScanner, { called }] = useMutation(SCAN_MUTATION)

  return (
    <div>
      <SectionTitle nospace>
        {t('settings.scanner.title', 'Scanner')}
      </SectionTitle>
      <InputLabelDescription>
        {t(
          'settings.scanner.description',
          'Will scan all users for new or updated media'
        )}
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
        {t('settings.scanner.scan_all_users', 'Scan all users')}
      </Button>
      <PeriodicScanner />
      <ScannerConcurrentWorkers />
    </div>
  )
}

export default ScannerSection
