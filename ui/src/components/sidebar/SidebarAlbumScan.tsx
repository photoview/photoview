import React, { useState } from 'react'
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
  viewerCanUpload: boolean
}

export const SidebarAlbumScan = ({
  id,
  viewerCanUpload,
}: SidebarAlbumScanProps) => {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()

  const [buttonDisabled, setButtonDisabled] = useState(false)

  const [scanAlbum, { called }] = useMutation<scanAlbum, scanAlbumVariables>(
    SCAN_ALBUM_MUTATION,
    {
      variables: { albumId: id },
      // Without this, a rejected mutation is an unhandled promise
      // rejection and leaves the button disabled with no way to retry.
      onError: () => setButtonDisabled(false),
    }
  )

  if (!isAdmin && !viewerCanUpload) {
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
