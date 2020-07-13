import React from 'react'
import PropTypes from 'prop-types'

const getProtectedUrl = url => {
  const imgUrl = new URL(url)

  if (localStorage.getItem('token') == null) {
    // Get share token if not authorized

    const tokenRegex = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
    if (tokenRegex) {
      const token = tokenRegex[1]
      imgUrl.searchParams.set('token', token)
    }
  }

  return imgUrl.href
}

/**
 * An image that needs authorization to load
 */
export const ProtectedImage = ({ src, ...props }) => {
  // const [imgSrc, setImgSrc] = useState(null)

  // useEffect(() => {
  //   if (imageCache[src]) return

  //   const fetchController = new AbortController()
  //   let canceled = false

  //   setImgSrc('')

  //   const imgUrl = new URL(src)
  //   const fetchHeaders = {}

  //   if (localStorage.getItem('token') == null) {
  //     // Get share token if not authorized

  //     const tokenRegex = location.pathname.match(/^\/share\/([\d\w]+)(\/?.*)$/)
  //     if (tokenRegex) {
  //       const token = tokenRegex[1]
  //       imgUrl.searchParams.set('token', token)

  //       const tokenPassword = sessionStorage.getItem(`share-token-pw-${token}`)
  //       if (tokenPassword) {
  //         fetchHeaders['TokenPassword'] = tokenPassword
  //       }
  //     }
  //   }

  //   fetchProtectedImage(imgUrl.href, {
  //     signal: fetchController.signal,
  //     headers: fetchHeaders,
  //   })
  //     .then(newSrc => {
  //       if (!canceled) {
  //         setImgSrc(newSrc)
  //       }
  //     })
  //     .catch(error => {
  //       console.log('Fetch image error', error.message)
  //     })

  //   return function cleanup() {
  //     canceled = true
  //     fetchController.abort()
  //   }
  // }, [src])

  return (
    <img {...props} src={getProtectedUrl(src)} crossOrigin="use-credentials" />
  )
}

ProtectedImage.propTypes = {
  src: PropTypes.string.isRequired,
}

export const ProtectedVideo = ({ media, ...props }) => {
  return (
    <video
      {...props}
      controls
      key={media.id}
      crossOrigin="use-credentials"
      poster={getProtectedUrl(media.thumbnail.url)}
    >
      <source src={getProtectedUrl(media.videoWeb.url)} type="video/mp4" />
    </video>
  )
}

ProtectedVideo.propTypes = {
  media: PropTypes.object.isRequired,
}
