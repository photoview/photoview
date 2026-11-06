import React, { useEffect, useState } from 'react'
import { useQuery, gql, useMutation } from '@apollo/client'
import { useForm } from 'react-hook-form'
import { INITIAL_SETUP_QUERY, login } from './loginUtilities'
import { authToken } from '../../helpers/authentication'
import { API_ENDPOINT } from '../../apolloClient'
import urlJoin from 'url-join'

import { useTranslation } from 'react-i18next'
import { Helmet } from 'react-helmet'
import { useNavigate } from 'react-router'
import { TextField } from '../../primitives/form/Input'
import MessageBox from '../../primitives/form/MessageBox'
import { CheckInitialSetup } from './__generated__/CheckInitialSetup'
import { Authorize, AuthorizeVariables } from './__generated__/Authorize'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

const LogoHeader = () => {
  const { t } = useTranslation()

  return (
    <div className="flex justify-center flex-col mb-14 mt-20">
      <img
        className="h-24"
        src={import.meta.env.BASE_URL + 'photoview-logo.svg'}
        alt="photoview logo"
      />
      <h1 className="text-3xl text-center mt-4">
        {t('login_page.welcome', 'Welcome to Photoview')}
      </h1>
    </div>
  )
}

const LoginForm = () => {
  const { t } = useTranslation()
  const {
    register,
    handleSubmit,
    formState: { errors: formErrors },
  } = useForm<LoginInputs>()

  const [authorize, { loading, data }] = useMutation<
    Authorize,
    AuthorizeVariables
  >(authorizeMutation, {
    onCompleted: data => {
      const { success, token } = data.authorizeUser

      if (success && token) {
        login(token)
      }
    },
  })

  const onSubmit = (data: LoginInputs) => {
    authorize({
      variables: {
        username: data.username,
        password: data.password,
      },
    })
  }

  const errorMessage =
    data && !data.authorizeUser.success ? data.authorizeUser.status : null

  return (
    <form
      className="mx-auto max-w-[500px] px-4"
      onSubmit={handleSubmit(onSubmit)}
      // loading={loading || (data && data.authorizeUser.success)}
    >
      <TextField
        sizeVariant="big"
        wrapperClassName="my-6"
        className="w-full"
        label={t('login_page.field.username', 'Username')}
        {...register('username', { required: true })}
        error={
          formErrors.username?.type == 'required'
            ? 'Please enter a username'
            : undefined
        }
      />
      <TextField
        sizeVariant="big"
        wrapperClassName="my-6"
        className="w-full"
        type="password"
        label={t('login_page.field.password', 'Password')}
        {...register('password')}
      />
      <input
        type="submit"
        disabled={loading}
        value={t('login_page.field.submit', 'Sign in')}
        className="rounded-md px-8 py-2 mt-2 focus:outline-none cursor-pointer bg-gradient-to-bl from-[#FF8246] to-[#D6264D] text-white font-semibold focus:ring-2 focus:ring-red-200 disabled:cursor-default disabled:opacity-80"
      />
      <MessageBox
        message={errorMessage}
        show={!!errorMessage}
        type="negative"
      />
    </form>
  )
}

type LoginInputs = {
  username: string
  password: string
}

const LoginPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [oidcEnabled, setOidcEnabled] = useState(false)
  const oidcError = new URLSearchParams(window.location.search).get(
    'oidc_error'
  )
  const oidcErrorMessage = oidcError
    ? oidcError === 'invalid_state'
      ? t(
          'login_page.oidc_invalid_state',
          'Your SSO sign-in expired. Please try again.'
        )
      : oidcError === 'not_linked'
      ? t(
          'login_page.oidc_not_linked',
          'This SSO account is not linked to a Photoview user.'
        )
      : t(
          'login_page.oidc_failed',
          'SSO sign-in failed. Please try again or use your password.'
        )
    : null

  useEffect(() => {
    if (typeof fetch !== 'function') return
    fetch(urlJoin(API_ENDPOINT, '/auth/oidc/status'), {
      credentials: 'include',
    })
      .then(response => (response.ok ? response.json() : null))
      .then((status: unknown) =>
        setOidcEnabled(
          typeof status === 'object' &&
            status !== null &&
            'enabled' in status &&
            status.enabled === true
        )
      )
      .catch(() => setOidcEnabled(false))
  }, [])

  const { data: initialSetupData } = useQuery<CheckInitialSetup>(
    INITIAL_SETUP_QUERY,
    { variables: {} }
  )

  useEffect(() => {
    if (authToken()) navigate('/')
  }, [])

  useEffect(() => {
    if (initialSetupData?.siteInfo?.initialSetup) navigate('/initialSetup')
  }, [initialSetupData?.siteInfo?.initialSetup])

  if (authToken() || initialSetupData?.siteInfo?.initialSetup) {
    return null
  }

  return (
    <>
      <Helmet>
        <title>{t('title.login', 'Login')} - Photoview</title>
      </Helmet>
      <div>
        <LogoHeader />
        <div className="mx-auto max-w-[500px]">
          <MessageBox
            message={oidcErrorMessage}
            show={!!oidcErrorMessage}
            type="negative"
          />
        </div>
        <LoginForm />
        {oidcEnabled && (
          <div className="mx-auto max-w-[500px] mt-6">
            <a
              href={urlJoin(API_ENDPOINT, '/auth/oidc/start')}
              className="inline-block rounded-md px-8 py-2 border border-gray-300 font-semibold"
            >
              {t('login_page.oidc_sign_in', 'Sign in with SSO')}
            </a>
          </div>
        )}
      </div>
    </>
  )
}

export default LoginPage
