import React from 'react'
import { LogoutButton } from '../../../Pages/SettingsPage/UserPreferences'

type ReDetectModalContentProps = {
  title: string
  children: React.ReactNode
  onClose: () => void
  onConfirm: () => void
  showLogoutOnly?: boolean
}

const ReDetectModalContent: React.FC<ReDetectModalContentProps> = ({
  title,
  children,
  onClose,
  onConfirm,
  showLogoutOnly,
}) => {
  return (
    <div className="fixed inset-0 flex items-center justify-center z-50">
      <div className="bg-black opacity-50 absolute inset-0"></div>
      <div
        className="bg-white dark:bg-gray-800 p-6 rounded shadow-lg z-10"
        style={{ whiteSpace: 'pre-wrap', maxWidth: '600px' }}
      >
        <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-gray-300">
          {title}
        </h2>
        {children}
        <div className="flex justify-end">
          {showLogoutOnly ? (
            <LogoutButton />
          ) : (
            <>
              <button
                type="button"
                className="mr-2 px-4 py-2 bg-gray-300 dark:bg-gray-600 text-gray-900 dark:text-gray-300 rounded"
                onClick={onClose}
              >
                Cancel
              </button>
              <button
                type="button"
                className="px-4 py-2 bg-blue-500 dark:bg-blue-700 text-white rounded"
                onClick={onConfirm}
              >
                Confirm
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default ReDetectModalContent
