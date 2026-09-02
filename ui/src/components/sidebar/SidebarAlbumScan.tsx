import React, { useEffect, useState } from 'react'
import { useMutation, gql } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { SidebarSection, SidebarSectionTitle } from './SidebarComponents'
import { useIsAdmin } from '../routes/AuthorizedRoute'
import { scanAlbum, scanAlbumVariables } from './__generated__/scanAlbum'

const SCAN_ALBUM_MUTATION = gql`
  mutation scanAlbum($albumId: ID!) {
    scanAlbum(albumId: $albumId) {
      success
    }
  }
`

type SidebarAlbumScanProps = {
  id: string
}

export const SidebarAlbumScan = ({ id }: SidebarAlbumScanProps) => {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()

  const [scanAlbum, { called }] = useMutation<scanAlbum, scanAlbumVariables>(
    SCAN_ALBUM_MUTATION,
    {
      variables: { albumId: id },
    }
  )

  const [buttonDisabled, setButtonDisabled] = useState(false)

  useEffect(() => {
    setButtonDisabled(false)
  }, [id])

  if (!isAdmin) {
    return null
  }

  return (
    <SidebarSection>
      <SidebarSectionTitle>
        {t('sidebar.album.scan.title', 'Scanner')}
      </SidebarSectionTitle>
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 dark:border-dark-border2 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="disabled:opacity-50 text-green-500 font-bold uppercase text-xs"
                  disabled={buttonDisabled || called}
                  onClick={() => {
                    setButtonDisabled(true)
                    scanAlbum()
                  }}
                >
                  <span>
                    {t(
                      'sidebar.album.scan.rescan',
                      'Rescan this album and its sub-albums'
                    )}
                  </span>
                </button>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </SidebarSection>
  )
}

export default SidebarAlbumScan
