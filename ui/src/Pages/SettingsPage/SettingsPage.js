import React from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'

import Layout from '../../Layout'

import ScannerSection from './ScannerSection'
import UsersTable from './Users/UsersTable'

export const SectionTitle = styled.h2`
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

  return (
    <Layout title={t('title.settings', 'Settings')}>
      <ScannerSection />
      <UsersTable />
    </Layout>
  )
}

export default SettingsPage
