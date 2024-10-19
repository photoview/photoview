import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useForm } from 'react-hook-form'
import { useMutation, gql } from '@apollo/client'
import { TextField } from '../../../primitives/form/Input'
import {
  clearTokenCookie,
  saveTokenCookie,
} from '../../../helpers/authentication'
import { MessageState } from '../../messages/Messages'
import { NotificationType } from '../../../__generated__/globalTypes'
import { SIDEBAR_MEDIA_QUERY } from '../MediaSidebar/MediaSidebar'
import { useModal } from './ReDetectFacesContext'
import ReDetectModalContent from './ReDetectModalContent'

type ReDetectFacesModalProps = {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
}

type ReenterPasswordResponse = {
  reenterPassword: {
    success: boolean
    token: string
  }
}

const RE_DETECT_FACES_MUTATION = gql`
  mutation reDetectFaces($mediaId: ID!) {
    reDetectFaces(mediaId: $mediaId)
  }
`

const REENTER_PASSWORD_MUTATION = gql`
  mutation reenterPassword($password: String!) {
    reenterPassword(password: $password) {
      success
      token
    }
  }
`

const ReDetectFacesModal: React.FC<ReDetectFacesModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
}) => {
  const { t } = useTranslation()
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<{ password: string }>()
  const { modalParams } = useModal()
  const mediaId = modalParams?.mediaId || 'unknown'
  const [attempts, setAttempts] = useState(0)
  const [passwordError, setPasswordError] = useState<string | undefined>(
    undefined
  )

  const [reDetectFaces] = useMutation(RE_DETECT_FACES_MUTATION, {
    refetchQueries: [
      { query: SIDEBAR_MEDIA_QUERY, variables: { id: mediaId } },
    ],
    onCompleted: () => {
      MessageState.add({
        key: `reDetectFaces-${mediaId}`,
        type: NotificationType.Message,
        props: {
          header: t('sidebar.album.redetect_faces_success', 'Success'),
          content: t(
            'sidebar.album.redetect_faces_success',
            'Faces re-detected successfully'
          ),
        },
      })
      onConfirm()
    },
    onError: error => {
      MessageState.add({
        key: `reDetectFaces-${mediaId}-error`,
        type: NotificationType.Message,
        props: {
          header: t('sidebar.album.redetect_faces_error', 'Error'),
          content:
            t(
              'sidebar.album.redetect_faces_error',
              'Error re-detecting faces: '
            ) + error.message,
        },
      })
      onConfirm()
    },
  })

  const [reenterPassword] = useMutation<ReenterPasswordResponse>(
    REENTER_PASSWORD_MUTATION,
    {
      onCompleted: data => {
        const { success, token } = data.reenterPassword

        if (success && token) {
          saveTokenCookie(token)
          handleConfirm()
        } else {
          handleIncorrectPassword()
        }
        reset()
      },
      onError: () => {
        handleIncorrectPassword()
        reset()
      },
    }
  )

  const handleConfirm = () => {
    reDetectFaces({
      variables: {
        mediaId: mediaId,
      },
    })
  }

  const handleIncorrectPassword = () => {
    setAttempts(prev => {
      const newAttempts = prev + 1
      if (newAttempts >= 3) {
        // Log out the user in the background
        clearTokenCookie()
      } else {
        setPasswordError(t('modal.incorrect_password', 'Incorrect password'))
      }
      return newAttempts
    })
  }

  const onSubmit = (data: { password: string }) => {
    setPasswordError(undefined)
    reenterPassword({ variables: { password: data.password } })
  }

  const handleClose = () => {
    setPasswordError(undefined)
    setAttempts(0)
    reset()
    onClose()
  }

  if (!isOpen) return null

  if (attempts >= 3) {
    return (
      <ReDetectModalContent
        title={t('sidebar.album.incorrect_password', 'Incorrect Password')}
        onClose={handleClose}
        onConfirm={() => null}
        showLogoutOnly={true}
      >
        <p className="mb-4 text-gray-700 dark:text-gray-400">
          {t(
            'sidebar.album.incorrect_password_message',
            'You have entered the incorrect password 3 times. You have been logged out.'
          )}
        </p>
      </ReDetectModalContent>
    )
  }

  return (
    <ReDetectModalContent
      title={t('sidebar.album.confirm_redetect_faces', 'Confirm Re-detection')}
      onClose={handleClose}
      onConfirm={handleSubmit(onSubmit)}
    >
      <p className="mb-4 text-gray-700 dark:text-gray-400">
        {t(
          'sidebar.album.redetect_faces_warning',
          'Are you sure you want to re-detect faces? This action will clear all unlabeled faces for the current media (they are selected in a circle). Then, the newly-detected faces will be auto-merged into the best matching groups, which might not be the current ones.'
        )}
      </p>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="mb-4">
          <TextField
            sizeVariant="big"
            wrapperClassName="my-6"
            className="w-full"
            type="password"
            label={t('login_page.field.password', 'Re-enter your password')}
            {...register('password', { required: true })}
            error={
              errors.password?.type === 'required'
                ? t('modal.password_required', 'Password is required')
                : passwordError
            }
          />
        </div>
      </form>
    </ReDetectModalContent>
  )
}

export default ReDetectFacesModal
