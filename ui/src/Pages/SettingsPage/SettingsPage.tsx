import React from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from 'semantic-ui-react'
import styled from 'styled-components'
import { useIsAdmin } from '../../components/routes/AuthorizedRoute'

import Layout from '../../components/layout/Layout'

import ScannerSection from './ScannerSection'
import UserPreferences from './UserPreferences'
import UsersTable from './Users/UsersTable'
import VersionInfo from './VersionInfo'

export const SectionTitle = styled.h2<{ nospace?: boolean }>`
  margin-top: ${({ nospace }) => (nospace ? '0' : '1.4em')} !important;
  padding-bottom: 0.3em;
  border-bottom: 1px solid #ddd;
`

export const InputLabelTitle = styled.p`
  font-size: 1.1em;
  font-weight: 600;
  margin: 1em 0 0 !important;
`

export const InputLabelDescription = styled.p`
  font-size: 0.9em;
  margin: 0 0 0.5em !important;
`

const SettingsPage = () => {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()

  return (
    <Layout title={t('title.settings', 'Settings')}>
      <UserPreferences />
      {isAdmin && (
        <>
          <ScannerSection />
          <UsersTable />
        </>
      )}
      <Button
        style={{ marginTop: 24 }}
        onClick={() => {
          location.href = '/logout'
        }}
      >
        {t('settings.logout', 'Log out')}
      </Button>
      <VersionInfo />
    </Layout>
  )
}

export default SettingsPage
