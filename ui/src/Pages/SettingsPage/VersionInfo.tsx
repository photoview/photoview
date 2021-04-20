import React from 'react'
import styled from 'styled-components'
import { useTranslation } from 'react-i18next'
import {
  InputLabelDescription,
  InputLabelTitle,
  SectionTitle,
} from './SettingsPage'

const VERSION = process.env.VERSION ? process.env.VERSION : 'undefined'
const BUILD_DATE = process.env.BUILD_DATE ? process.env.BUILD_DATE : 'undefined'

const VersionInfoWrapper = styled.div`
  margin-bottom: 24px;
`

const VersionInfo = () => {
  const { t } = useTranslation()

  return (
    <VersionInfoWrapper>
      <SectionTitle>
        {t('settings.version_info.title', 'Photoview Version')}
      </SectionTitle>
      <InputLabelTitle>
        {t('settings.version_info.version_title', 'Release Version')}
      </InputLabelTitle>
      <InputLabelDescription>{VERSION}</InputLabelDescription>
      <InputLabelTitle>
        {t('settings.version_info.build_date_title', 'Build date')}
      </InputLabelTitle>
      <InputLabelDescription>{BUILD_DATE}</InputLabelDescription>
    </VersionInfoWrapper>
  )
}

export default VersionInfo
