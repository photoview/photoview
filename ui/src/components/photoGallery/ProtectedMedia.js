import React from 'react'
import PropTypes from 'prop-types'

const isNativeLazyLoadSupported = 'loading' in HTMLImageElement.prototype
const placeholder = 'data:image/gif;base64,R0lGODlhAQABAPAAAAAAAAAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=='

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
 * Set lazyLoading to true if you want the image to be loaded once it enters the viewport
 * Native lazy load via HTMLImageElement.loading attribute will be preferred if it is supported by the browser,
 * otherwise IntersectionObserver will be used.
 */
export const ProtectedImage = ({ src, lazyLoading, ...props }) => {
  if (!isNativeLazyLoadSupported && lazyLoading) {
    props['data-src'] = getProtectedUrl(src)
  }

  if (isNativeLazyLoadSupported && lazyLoading) {
    props.loading = 'lazy'
  }

  return (
    <img
      key={src}
      {...props}
      src={
        lazyLoading && !isNativeLazyLoadSupported
          ? placeholder
          : getProtectedUrl(src)
      }
      crossOrigin="use-credentials"
    />
  )
}

ProtectedImage.propTypes = {
  src: PropTypes.string,
  lazyLoading: PropTypes.bool,
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
