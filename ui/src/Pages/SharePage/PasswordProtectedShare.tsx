import React, { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { TextField } from '../../primitives/form/Input'
import { MessageContainer } from './SharePage'
// import { Message, Header, Form, Input, Icon } from 'semantic-ui-react'

type ProtectedTokenEnterPasswordProps = {
  refetchWithPassword(password: string): void
  loading: boolean
}

const PasswordProtectedShare = ({
  refetchWithPassword,
  loading = false,
}: ProtectedTokenEnterPasswordProps) => {
  const { t } = useTranslation()

  const {
    register,
    watch,
    formState: { errors },
    handleSubmit,
  } = useForm()

  // const [passwordValue, setPasswordValue] = useState('')
  const [invalidPassword, setInvalidPassword] = useState(false)

  const onSubmit = () => {
    refetchWithPassword(watch('password'))
    setInvalidPassword(true)
  }

  let errorMessage = undefined
  if (invalidPassword && !loading) {
    errorMessage = t(
      'share_page.wrong_password',
      'Wrong password, please try again.'
    )
  } else if (errors.password?.type === 'required') {
    errorMessage = t(
      'share_page.protected_share.password_required_error',
      'Password is required'
    )
  }

  return (
    <MessageContainer>
      <h1 className="text-xl">
        {t('share_page.protected_share.title', 'Protected share')}
      </h1>
      <p className="mb-4">
        {t(
          'share_page.protected_share.description',
          'This share is protected with a password.'
        )}
      </p>
      <TextField
        {...register('password', { required: true })}
        label={t('login_page.field.password', 'Password')}
        type="password"
        loading={loading}
        disabled={loading}
        action={handleSubmit(onSubmit)}
        error={errorMessage}
        fullWidth={true}
        sizeVariant="big"
      />
      {/* <Form.Field>
          <label>{t('login_page.field.password', 'Password')}</label>
          <Input
            loading={loading}
            disabled={loading}
            onKeyUp={(event: KeyboardEvent) =>
              event.key == 'Enter' && onSubmit()
            }
            onChange={e => setPasswordValue(e.target.value)}
            placeholder={t('login_page.field.password', 'Password')}
            type="password"
            icon={<Icon onClick={onSubmit} link name="arrow right" />}
          />
        </Form.Field>
        {errorMessage} */}
    </MessageContainer>
  )
}

export default PasswordProtectedShare
