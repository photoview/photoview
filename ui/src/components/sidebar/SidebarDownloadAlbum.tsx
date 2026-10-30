import React from 'react'
import { useLocation } from 'react-router'
import { useTranslation } from 'react-i18next'
import { API_ENDPOINT } from '../../apolloClient'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import SidebarTable from './SidebarTable'
import { authToken } from '../../helpers/authentication'

type SidebarAlbumDownladProps = {
  albumID: string
}

export const generateDownloadUrl = (
  albumID: string,
  downloadType: {
    title: string
    description: string
    purpose: string
  },
  pathname = location.pathname
) => {
  const commonUrl = `${API_ENDPOINT}/download/album/${albumID}/${downloadType.purpose}`

  if (authToken() != null) {
    return commonUrl
  }

  const shareToken = pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
  if (shareToken == null) {
    return commonUrl
  }

  return `${commonUrl}?token=${shareToken[1]}`
}

const SidebarAlbumDownload = ({ albumID }: SidebarAlbumDownladProps) => {
  const { t } = useTranslation()
  const { pathname } = useLocation()

  const downloads = [
    {
      title: t('sidebar.album.download.thumbnails.title', 'Thumbnails'),
      description: t(
        'sidebar.album.download.thumbnails.description',
        'Low resolution images, no videos'
      ),
      purpose: 'thumbnail,video-thumbnail',
    },
    {
      title: t(
        'sidebar.album.download.high-resolutions.title',
        'High resolutions'
      ),
      description: t(
        'sidebar.album.download.high-resolutions.description',
        'High resolution jpegs of RAW images'
      ),
      purpose: 'high-res',
    },
    {
      title: t('sidebar.album.download.originals.title', 'Originals'),
      description: t(
        'sidebar.album.download.originals.description',
        'The original images and videos'
      ),
      purpose: 'original',
    },
    {
      title: t('sidebar.album.download.web-videos.title', 'Converted videos'),
      description: t(
        'sidebar.album.download.web-videos.description',
        'Videos that have been optimized for web'
      ),
      purpose: 'video-web',
    },
  ]

  const downloadRows = downloads.map(x => (
    <SidebarTable.Row
      key={x.purpose}
      onClick={() => {
        const url = generateDownloadUrl(albumID, x, pathname)
        return (location.href = url)
      }}
      tabIndex={0}
    >
      <td className="pl-4 py-2">{`${x.title}`}</td>
      <td className="pr-4 py-2 text-sm text-gray-800 dark:text-gray-400 italic">{`${x.description}`}</td>
    </SidebarTable.Row>
  ))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.download.title', 'Download')}
      </SidebarSectionTitle>

      <SidebarTable.Table>
        <SidebarTable.Head>
          <SidebarTable.HeadRow>
            <th className="px-4 py-2" colSpan={2}>
              {t('sidebar.download.table_columns.name', 'Name')}
            </th>
          </SidebarTable.HeadRow>
        </SidebarTable.Head>
        <tbody>{downloadRows}</tbody>
      </SidebarTable.Table>
    </SidebarSection>
  )
}

export default SidebarAlbumDownload
