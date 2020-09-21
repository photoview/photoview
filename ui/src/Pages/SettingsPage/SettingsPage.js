import React from 'react'
import styled from 'styled-components'

import Layout from '../../Layout'

import ScannerSection from './ScannerSection'
import UsersTable from './UsersTable'

export const SectionTitle = styled.h2`
  margin-top: ${({ nospace }) => (nospace ? '0' : '1.4em')} !important;
  padding-bottom: 0.3em;
  border-bottom: 1px solid #ddd;
`

const SettingsPage = () => {
  return (
    <Layout>
      <ScannerSection />
      <UsersTable />
    </Layout>
  )
}

export default SettingsPage
