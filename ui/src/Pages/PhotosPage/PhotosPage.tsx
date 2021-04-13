import React from 'react'
import Layout from '../../Layout'
import TimelineGallery from '../../components/timelineGallery/TimelineGallery'
import { useTranslation } from 'react-i18next'

const PhotosPage = () => {
  const { t } = useTranslation()

  return (
    <>
      <Layout title={t('photos_page.title', 'Photos')}>
        <TimelineGallery />
      </Layout>
    </>
  )
}

export default PhotosPage
