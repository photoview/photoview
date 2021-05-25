import React from 'react'
import { NavLink } from 'react-router-dom'
import { useQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'

export const MAPBOX_QUERY = gql`
  query mapboxEnabledQuery {
    mapboxToken
  }
`

type MenuButtonProps = {
  to: string
  exact: boolean
  label: string
  gradient: string
  activeClasses?: string
  className?: string
  icon?: React.ReactChild
}

const MenuButton = ({
  to,
  exact,
  label,
  gradient,
  icon,
  activeClasses,
  className,
}: MenuButtonProps) => {
  return (
    <NavLink
      to={to}
      exact={exact}
      className={`rounded-lg my-2 ${className}`}
      activeClassName={`ring-4 lg:ring-4 ${activeClasses}`}
    >
      <li className="flex items-center">
        <div
          className={`w-12 h-12 p-1.5 lg:w-8 lg:h-8 lg:p-1 bg-gradient-to-br ${gradient} w-full h-full rounded-lg`}
        >
          {icon}
        </div>
        <span className="hidden lg:block ml-2">{label}</span>
      </li>
    </NavLink>
  )
}

const MenuSeparator = () => (
  <hr className="hidden lg:block my-3 border-gray-200" />
)

export const MainMenu = () => {
  const { t } = useTranslation()

  const mapboxQuery = authToken() ? useQuery(MAPBOX_QUERY) : null

  const mapboxEnabled = !!mapboxQuery?.data?.mapboxToken

  return (
    <div className="absolute lg:relative w-full bottom-0 z-10 bg-white shadow-separator lg:shadow-none lg:w-[240px] lg:ml-8 lg:mr-5 lg:my-4">
      <ul className="flex justify-around py-2 px-2 z-10 max-w-lg mx-auto lg:flex-col lg:p-0">
        <MenuButton
          to="/photos"
          exact
          label={t('sidemenu.photos', 'Timeline')}
          gradient="from-[#AAD4F8] to-[#80B2E8]"
          activeClasses="ring-[#f1f8ff] bg-[#f1f8ff]"
          className="outline-none focus:ring-2 focus:ring-blue-200 focus:ring-offset-2"
          icon={
            <svg viewBox="0 0 24 24" fill="white">
              <path d="M15.6971052,10 L23.9367603,21.526562 C23.6878671,22.9323278 22.4600272,24 20.982819,24 L5.851819,24 L15.6971052,10 Z" />
              <path
                d="M5.59307375,14 L15.562,24 L2.982819,24 C1.43507633,24 0.161084327,22.8279341 -3.56225466e-14,21.3229592 L5.59307375,14 Z"
                fillOpacity="0.75"
              />
            </svg>
          }
        />
        <MenuButton
          to="/albums"
          exact
          label={t('sidemenu.albums', 'Albums')}
          gradient="from-[#F8AAAA] to-[#E88380]"
          activeClasses="ring-[#fff1f2] bg-[#fff1f2]"
          className="outline-none focus:ring-2 focus:ring-red-200 focus:ring-offset-2"
          icon={
            <svg viewBox="0 0 24 24" fill="white">
              <path
                d="M6,16 L19,16 C19.5522847,16 20,16.4477153 20,17 L20,21 C20,21.5522847 19.5522847,22 19,22 L6,22 C4.8954305,22 4,21.1045695 4,20 L4,18 C4,16.8954305 4.8954305,16 6,16 Z"
                fillOpacity="0.75"
              />
              <path d="M19,2 C19.5522847,2 20,2.44771525 20,3 L20,19 C20,19.5522847 19.5522847,20 19,20 L6,20 C4.8954305,20 4,19.1045695 4,18 L4,4 C4,2.8954305 4.8954305,2 6,2 L19,2 Z M14.4676845,9 L11.5079767,12.9536523 L9.50029382,10.8745763 L7,14 L18,14 L14.4676845,9 Z M10.75,9 C10.3357864,9 10,9.33578644 10,9.75 C10,10.1642136 10.3357864,10.5 10.75,10.5 C11.1642136,10.5 11.5,10.1642136 11.5,9.75 C11.5,9.33578644 11.1642136,9 10.75,9 Z" />
            </svg>
          }
        />
        {mapboxEnabled ? (
          <MenuButton
            to="/places"
            exact
            label={t('sidemenu.places', 'Places')}
            gradient="from-[#B8EF7F] to-[#8CD77B]"
            activeClasses="ring-[#e3fee5] bg-[#e3fee5]"
            className="outline-none focus:ring-2 focus:ring-green-100 focus:ring-offset-2"
            icon={
              <svg viewBox="0 0 24 24" fill="white">
                <path d="M3.4,3.34740684 C3.47896999,3.34740684 3.55617307,3.37078205 3.62188008,3.41458672 L9,7 L9,21 L3.4452998,17.2968665 C3.16710114,17.1114008 3,16.7991694 3,16.4648162 L3,3.74740684 C3,3.52649294 3.1790861,3.34740684 3.4,3.34740684 Z M15,3 L21.4961389,6.71207939 C21.8077139,6.89012225 22,7.22146569 22,7.58032254 L22,20.3107281 C22,20.531642 21.8209139,20.7107281 21.6,20.7107281 C21.5303892,20.7107281 21.4619835,20.692562 21.4015444,20.6580254 L15,17 L15,3 Z" />
                <polygon
                  fillOpacity="0.75"
                  transform="translate(12, 12) scale(1, -1) translate(-12, -12)"
                  points="9 3 15 7 15 21 9 17"
                />
              </svg>
            }
          />
        ) : null}
        <MenuButton
          to="/people"
          exact
          label={t('sidemenu.people', 'People')}
          gradient="from-[#F6F16E] to-[#F3C688]"
          activeClasses="ring-[#fff7e4] bg-[#fff7e4]"
          className="outline-none focus:ring-2 focus:ring-yellow-100 focus:ring-offset-2"
          icon={
            <svg viewBox="0 0 24 24" fill="white">
              <path
                d="M12,15 C15.1826579,15 18.0180525,16.4868108 19.8494955,18.8037439 L20,19 C20,20.6568542 18.6568542,22 17,22 L7,22 C5.34314575,22 4,20.6568542 4,19 L4.15050454,18.8037439 C5.9819475,16.4868108 8.81734212,15 12,15 Z"
                fillOpacity="0.75"
              ></path>
              <circle cx="12" cy="11" r="6"></circle>
            </svg>
          }
        />
        <MenuSeparator />
        <MenuButton
          to="/settings"
          exact
          label={t('sidemenu.settings', 'Settings')}
          gradient="from-[#C7E2E2] to-[#96AFBA]"
          activeClasses="ring-[#e4f0f8] bg-[#e4f0f8]"
          className="outline-none focus:ring-2 focus:ring-gray-200 focus:ring-offset-2"
          icon={
            <svg viewBox="0 0 24 24" fill="white">
              <path
                d="M13.4691754,16.7806702 L13,21 L11,21 L10.5318353,16.7809803 C10.9960818,16.9233714 11.4890921,17 12,17 C12.5112786,17 13.0046337,16.9232601 13.4691754,16.7806702 Z M16.4192861,14.3409153 L19.0710678,17.6568542 L17.6568542,19.0710678 L14.3409153,16.4192861 C15.2243208,15.9503691 15.9503691,15.2243208 16.4192861,14.3409153 Z M9.65908474,16.4192861 L6.34314575,19.0710678 L4.92893219,17.6568542 L7.5807139,14.3409153 C8.04963086,15.2243208 8.77567918,15.9503691 9.65908474,16.4192861 Z M7,12 C7,12.5112786 7.07673988,13.0046337 7.21932976,13.4691754 L3,13 L3,11 L7.21901966,10.5318353 C7.07662862,10.9960818 7,11.4890921 7,12 Z M16.7809803,10.5318353 L21,11 L21,13 L16.7806702,13.4691754 C16.9232601,13.0046337 17,12.5112786 17,12 C17,11.4890921 16.9233714,10.9960818 16.7809803,10.5318353 Z M6.34314575,4.92893219 L9.65908474,7.5807139 C8.77567918,8.04963086 8.04963086,8.77567918 7.5807139,9.65908474 L4.92893219,6.34314575 L6.34314575,4.92893219 Z M17.6568542,4.92893219 L19.0710678,6.34314575 L16.4192861,9.65908474 C15.9503691,8.77567918 15.2243208,8.04963086 14.3409153,7.5807139 L17.6568542,4.92893219 Z M13,3 L13.4691754,7.21932976 C13.0046337,7.07673988 12.5112786,7 12,7 C11.4890921,7 10.9960818,7.07662862 10.5318353,7.21901966 L11,3 L13,3 Z"
                fillOpacity="0.76"
              />
              <path d="M12,5 C15.8659932,5 19,8.13400675 19,12 C19,15.8659932 15.8659932,19 12,19 C8.13400675,19 5,15.8659932 5,12 C5,8.13400675 8.13400675,5 12,5 Z M12,8 C9.790861,8 8,9.790861 8,12 C8,14.209139 9.790861,16 12,16 C14.209139,16 16,14.209139 16,12 C16,9.790861 14.209139,8 12,8 Z" />
            </svg>
          }
        />
      </ul>
    </div>
  )
}

export default MainMenu
