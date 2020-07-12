import React, { useEffect, useState } from 'react'
import PropTypes from 'prop-types'

// let imageCache = {}

// export async function fetchProtectedImage(
//   src,
//   { signal, headers: customHeaders } = { signal: null, headers: null }
// ) {
//   if (src) {
//     if (imageCache[src]) {
//       return imageCache[src]
//     }

//     let headers = {
//       ...customHeaders,
//     }
//     if (localStorage.getItem('token')) {
//       headers['Authorization'] = `Bearer ${localStorage.getItem('token')}`
//     }

//     let image = await fetch(src, {
//       headers,
//       signal,
//     })

//     image = await image.blob()
//     const url = URL.createObjectURL(image)

//     // eslint-disable-next-line require-atomic-updates
//     imageCache[src] = url

//     return url
//   }
// }

/**
 * An image that needs a authorization header to load
 */
const ProtectedImage = ({ src, ...props }) => {
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

  return <img {...props} src={src} crossOrigin="use-credentials" />
}

ProtectedImage.propTypes = {
  src: PropTypes.string,
}

export default ProtectedImage
