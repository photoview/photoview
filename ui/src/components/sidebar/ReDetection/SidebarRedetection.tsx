import React from 'react'
import { useTranslation } from 'react-i18next'
import { useModal } from './ReDetectFacesContext'

type SidebarReDetectionProps = {
  mediaId: string
  setAlwaysShowFaces: (value: boolean) => void
}

export const SidebarReDetection = ({
  mediaId,
  setAlwaysShowFaces,
}: SidebarReDetectionProps) => {
  const { t } = useTranslation()
  const { openModal } = useModal()

  const handleOpenModal = () => {
    setAlwaysShowFaces(true)
    openModal({ mediaId }, () => setAlwaysShowFaces(false))
  }

  return (
    <div className="mt-4">
      <div>
        <table className="border-collapse w-full">
          <tfoot>
            <tr className="text-left border-gray-100 dark:border-dark-border2 border-b border-t">
              <td colSpan={2} className="pl-4 py-2">
                <button
                  className="disabled:opacity-50 font-bold uppercase text-xs"
                  style={{ color: '#FFA500' }} // orange
                  onClick={handleOpenModal}
                >
                  <span>
                    {t(
                      'sidebar.album.redetect_faces',
                      'Re-detect unlabeled faces'
                    )}
                  </span>
                </button>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  )
}
