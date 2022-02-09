import React from 'react'
import { useTranslation } from 'react-i18next'
import styled from 'styled-components'
import { useIsAdmin } from '../../components/routes/AuthorizedRoute'
import Layout from '../../components/layout/Layout'
import ScannerSection from './ScannerSection'
import UserPreferences from './UserPreferences'
import UsersTable from './Users/UsersTable'
import VersionInfo from './VersionInfo'
import classNames from 'classnames'

type SectionTitleProps = {
  children: string
  nospace?: boolean
}

export const SectionTitle = ({ children, nospace }: SectionTitleProps) => {
  return (
    <h2
      className={classNames(
        'pb-1 border-b border-gray-200 dark:border-dark-border text-xl mb-5',
        !nospace && 'mt-6'
      )}
    >
      {children}
    </h2>
  )
}

export const InputLabelTitle = styled.h3.attrs({
  className: 'font-semibold mt-4',
})``

export const InputLabelDescription = styled.p.attrs({
  className: 'text-sm mb-2',
})``

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
      <VersionInfo />
    </Layout>
  )
}

export default SettingsPage
