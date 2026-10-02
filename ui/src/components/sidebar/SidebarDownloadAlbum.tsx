import React from 'react'
import { useTranslation } from 'react-i18next'
import { API_ENDPOINT } from '../../apolloClient'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import SidebarTable from './SidebarTable'

type SidebarAlbumDownladProps = {
  albumID: string
}

const SidebarAlbumDownload = ({ albumID }: SidebarAlbumDownladProps) => {
  const { t } = useTranslation()

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

  const downloadRows = downloads.map(x => {
    const downloadLabel = t(
      'sidebar.download.actions.download',
      'Download {{name}}',
      { name: x.title }
    )

    return (
      <SidebarTable.Row key={x.purpose}>
        <td className="p-0">
          <SidebarTable.RowButton
            aria-label={downloadLabel}
            title={downloadLabel}
            onClick={() =>
              (location.href = `${API_ENDPOINT}/download/album/${albumID}/${x.purpose}`)
            }
          >
            {x.title}
          </SidebarTable.RowButton>
        </td>
        <td className="pr-4 py-2 text-sm text-gray-800 dark:text-gray-400 italic">{`${x.description}`}</td>
      </SidebarTable.Row>
    )
  })

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
