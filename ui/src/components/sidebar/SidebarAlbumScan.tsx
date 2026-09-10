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

  const [scanAlbum, { called, reset }] = useMutation<
    scanAlbum,
    scanAlbumVariables
  >(SCAN_ALBUM_MUTATION, {
    variables: { albumId: id },
    // `called` never resets on its own once the mutation has fired once,
    // and the button's disabled state below also depends on it - without
    // calling reset() here too, the button would stay disabled forever
    // after the very first click, success or failure alike.
    onCompleted: () => {
      setButtonDisabled(false)
      reset()
    },
    onError: () => {
      setButtonDisabled(false)
      reset()
    },
  })

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
