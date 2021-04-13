import React, { ReactChild } from 'react'
import PropTypes from 'prop-types'
import styled from 'styled-components'
import { NavLink } from 'react-router-dom'
import { Icon } from 'semantic-ui-react'
import Sidebar from './components/sidebar/Sidebar'
import { useQuery, gql } from '@apollo/client'
import { Authorized } from './components/routes/AuthorizedRoute'
import { Helmet } from 'react-helmet'
import Header from './components/header/Header'
import { authToken } from './helpers/authentication'
import { useTranslation } from 'react-i18next'

export const ADMIN_QUERY = gql`
  query adminQuery {
    myUser {
      admin
    }
  }
`

export const MAPBOX_QUERY = gql`
  query mapboxEnabledQuery {
    mapboxToken
  }
`

const Container = styled.div`
  height: 100%;
  display: flex;
  overflow: hidden;
  position: relative;
`

const SideMenuContainer = styled.div`
  height: 100%;
  width: 80px;
  left: 0;
  padding-top: 70px;

  @media (max-width: 1000px) {
    width: 100%;
    height: 80px;
    position: fixed;
    background: white;
    z-index: 10;
    padding-top: 0;
    display: flex;
    bottom: 0;
    box-shadow: 0 0 2px rgba(0, 0, 0, 0.3);
  }
`

const Content = styled.div`
  margin-top: 70px;
  padding: 10px 12px 0;
  width: 100%;
  overflow-y: scroll;
`

const SideButtonLink = styled(NavLink)`
  text-align: center;
  padding-top: 8px;
  padding-left: 2px;
  display: block;
  width: 60px;
  height: 60px;
  margin: 10px;
  margin-bottom: 24px;

  font-size: 28px;

  color: #888;

  transition: transform 200ms, box-shadow 200ms;

  :hover {
    transform: scale(1.02);
  }
`

type SideButtonProps = {
  children: ReactChild | ReactChild[]
  to: string
  exact: boolean
}

const SideButton = ({
  children,
  to,
  exact,
  ...otherProps
}: SideButtonProps) => {
  return (
    <SideButtonLink
      {...otherProps}
      to={to}
      exact={exact}
      activeStyle={{ color: '#4183c4' }}
    >
      {children}
    </SideButtonLink>
  )
}

const SideButtonLabel = styled.div`
  font-size: 16px;
`

export const SideMenu = () => {
  const { t } = useTranslation()

  const mapboxQuery = authToken() ? useQuery(MAPBOX_QUERY) : null

  const mapboxEnabled = !!mapboxQuery?.data?.mapboxToken

  return (
    <SideMenuContainer>
      <SideButton to="/photos" exact>
        <Icon name="image" />
        <SideButtonLabel>{t('sidemenu.photos', 'Photos')}</SideButtonLabel>
      </SideButton>
      <SideButton to="/albums" exact>
        <Icon name="images" />
        <SideButtonLabel>{t('sidemenu.albums', 'Albums')}</SideButtonLabel>
      </SideButton>
      {mapboxEnabled ? (
        <SideButton to="/places" exact>
          <Icon name="map" />
          <SideButtonLabel>{t('sidemenu.places', 'Places')}</SideButtonLabel>
        </SideButton>
      ) : null}
      <SideButton to="/people" exact>
        <Icon name="user" />
        <SideButtonLabel>{t('sidemenu.people', 'People')}</SideButtonLabel>
      </SideButton>
      <SideButton to="/settings" exact>
        <Icon name="settings" />
        <SideButtonLabel>{t('sidemenu.settings', 'Settings')}</SideButtonLabel>
      </SideButton>
    </SideMenuContainer>
  )
}

type LayoutProps = {
  children: React.ReactNode
  title: string
}

const Layout = ({ children, title, ...otherProps }: LayoutProps) => {
  return (
    <Container {...otherProps} data-testid="Layout">
      <Helmet>
        <title>{title ? `${title} - Photoview` : `Photoview`}</title>
      </Helmet>
      <Authorized>
        <SideMenu />
      </Authorized>
      <Sidebar>
        <Content id="layout-content">
          {children}
          <div style={{ height: 24 }}></div>
        </Content>
      </Sidebar>
      <Header />
    </Container>
  )
}

Layout.propTypes = {
  children: PropTypes.any.isRequired,
  title: PropTypes.string,
}

export default Layout
