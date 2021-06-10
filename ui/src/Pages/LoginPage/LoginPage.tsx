import React from 'react'
import { useQuery, gql, useMutation } from '@apollo/client'
import { useForm, SubmitHandler } from 'react-hook-form'
import { checkInitialSetupQuery, login } from './loginUtilities'
import { authToken } from '../../helpers/authentication'

import logoPath from '../../assets/photoview-logo.svg'
import { useTranslation } from 'react-i18next'
import { Helmet } from 'react-helmet'
import { Redirect } from 'react-router'
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
    <div className="flex justify-center flex-col pb-20 pt-12">
      <img className="h-24" src={logoPath} alt="photoview logo" />
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
  } = useForm()

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

  const onSubmit: SubmitHandler<LoginInputs> = data => {
    authorize({
      variables: {
        username: data.username,
        password: data.password,
      },
    })
  }

  console.log('errors', formErrors)

  const errorMessage =
    data && !data.authorizeUser.success ? data.authorizeUser.status : null

  return (
    <form
      className="mx-auto max-w-[450px] px-4"
      onSubmit={handleSubmit(onSubmit)}
      // loading={loading || (data && data.authorizeUser.success)}
    >
      <TextField
        className="w-full my-4"
        label={t('login_page.field.username', 'Username')}
        {...register('username', { required: true })}
        error={
          formErrors.username?.type == 'required'
            ? 'Please enter a username'
            : undefined
        }
      />
      <TextField
        className="w-full my-4"
        type="password"
        label={t('login_page.field.password', 'Password')}
        {...register('password', { required: true })}
        error={
          formErrors.password?.type == 'required'
            ? 'Please enter a password'
            : undefined
        }
      />
      <input
        type="submit"
        disabled={loading}
        value={t('login_page.field.submit', 'Sign in') as string}
        className="rounded-md px-8 py-2 mt-1 focus:outline-none cursor-pointer bg-gradient-to-bl from-[#FF8246] to-[#D6264D] text-white font-semibold focus:ring-2 focus:ring-red-200 disabled:cursor-default disabled:opacity-80"
      />
      <MessageBox
        message={errorMessage}
        show={!!errorMessage}
        type="negative"
      />
    </form>
  )
}

// background-image: radial-gradient(0% 100%, #FF8246 0%, #D6264D 100%);
// border: 2px solid rgba(255,51,0,0.29);
// border-radius: 6px;

type LoginInputs = {
  username: string
  password: string
}

const LoginPage = () => {
  const { t } = useTranslation()

  const { data: initialSetupData } = useQuery<CheckInitialSetup>(
    checkInitialSetupQuery
  )

  if (authToken()) {
    return <Redirect to="/" />
  }

  return (
    <>
      <Helmet>
        <title>{t('title.login', 'Login')} - Photoview</title>
      </Helmet>
      {initialSetupData?.siteInfo?.initialSetup && (
        <Redirect to="/initialSetup" />
      )}
      <div>
        <LogoHeader />
        <LoginForm />
      </div>
    </>
  )
}

export default LoginPage
