import React from 'react'
import { useTranslation } from 'react-i18next'
import { API_ENDPOINT } from '../../apolloClient'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'

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

  const downloadRows = downloads.map(x => (
    <tr
      className="cursor-pointer border-gray-100 border-b hover:bg-gray-50 focus:bg-gray-50"
      key={x.purpose}
      onClick={() =>
        (location.href = `${API_ENDPOINT}/download/album/${albumID}/${x.purpose}`)
      }
      tabIndex={0}
    >
      <td className="pl-4 py-2">{`${x.title}`}</td>
      <td className="pr-4 py-2 text-sm text-gray-800 italic">{`${x.description}`}</td>
    </tr>
  ))

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.download.title', 'Download')}
      </SidebarSectionTitle>

      <table className="table-auto w-full">
        <thead className="bg-[#f9f9fb]">
          <tr className="text-left uppercase text-xs border-gray-100 border-b border-t">
            <th className="px-4 py-2" colSpan={2}>
              {t('sidebar.download.table_columns.name', 'Name')}
            </th>
          </tr>
        </thead>
        <tbody>{downloadRows}</tbody>
      </table>
    </SidebarSection>
  )
}

export default SidebarAlbumDownload
