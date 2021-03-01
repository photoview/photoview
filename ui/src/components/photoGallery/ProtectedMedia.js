import React from 'react'
import PropTypes from 'prop-types'

const getProtectedUrl = url => {
  if (url == null) return null

  const imgUrl = new URL(url, location.origin)

  const tokenRegex = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
  if (tokenRegex) {
    const token = tokenRegex[1]
    imgUrl.searchParams.set('token', token)
  }

  return imgUrl.href
}

/**
 * An image that needs authorization to load
 */
export const ProtectedImage = ({ src, ...props }) => (
  <img
    key={src}
    {...props}
    src={getProtectedUrl(src)}
    crossOrigin="use-credentials"
  />
)

ProtectedImage.propTypes = {
  src: PropTypes.string,
}

export const ProtectedVideo = ({ media, ...props }) => (
  <video
    {...props}
    controls
    key={media.id}
    crossOrigin="use-credentials"
    poster={getProtectedUrl(media.thumbnail?.url)}
  >
    <source src={getProtectedUrl(media.videoWeb.url)} type="video/mp4" />
  </video>
)

ProtectedVideo.propTypes = {
  media: PropTypes.object.isRequired,
}
