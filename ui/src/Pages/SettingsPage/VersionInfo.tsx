import React, { ReactElement } from 'react'
import styled from 'styled-components'
import { useTranslation } from 'react-i18next'
import {
  InputLabelDescription,
  InputLabelTitle,
  SectionTitle,
} from './SettingsPage'

const VERSION = import.meta.env.REACT_APP_BUILD_VERSION ?? 'undefined'
const BUILD_DATE = import.meta.env.REACT_APP_BUILD_DATE ?? 'undefined'

const COMMIT_SHA = import.meta.env.REACT_APP_BUILD_COMMIT_SHA
let commitLink: ReactElement

if (COMMIT_SHA) {
  commitLink = React.createElement(
    'a',
    {
      href: 'https://github.com/photoview/photoview/commit/' + COMMIT_SHA,
      title: COMMIT_SHA,
    },
    COMMIT_SHA.substring(0, 8)
  )
}

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
      <InputLabelDescription>
        {VERSION} ({commitLink})
      </InputLabelDescription>
      <InputLabelTitle>
        {t('settings.version_info.build_date_title', 'Build date')}
      </InputLabelTitle>
      <InputLabelDescription>{BUILD_DATE}</InputLabelDescription>
    </VersionInfoWrapper>
  )
}

export default VersionInfo
