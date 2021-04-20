import React from 'react'
import styled from 'styled-components'
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
  return (
    <VersionInfoWrapper>
      <SectionTitle>Photoview Version</SectionTitle>
      <InputLabelTitle>Release Version</InputLabelTitle>
      <InputLabelDescription>{VERSION}</InputLabelDescription>
      <InputLabelTitle>Build date</InputLabelTitle>
      <InputLabelDescription>{BUILD_DATE}</InputLabelDescription>
    </VersionInfoWrapper>
  )
}

export default VersionInfo
