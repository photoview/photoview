import React from 'react'
import Layout from '../../components/layout/Layout'
import { useTranslation } from 'react-i18next'
import TimelineGallery from '../../components/timelineGallery/TimelineGallery'

const TimelinePage = () => {
  const { t } = useTranslation()

  return (
    <>
      <Layout title={t('photos_page.title', 'Timeline')}>
        <TimelineGallery />
      </Layout>
    </>
  )
}

export default TimelinePage
