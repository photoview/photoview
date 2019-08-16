import React from 'react'
import PropTypes from 'prop-types'

let imageCache = {}

export async function fetchProtectedImage(src) {
  if (src) {
    if (imageCache[src]) {
      return imageCache[src]
    }

    let image = await fetch(src, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })

    image = await image.blob()
    const url = URL.createObjectURL(image)

    imageCache[src] = url

    return url
  }
}

/**
 * An image that needs a authorization header to load
 */
class ProtectedImage extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      imgSrc: null,
    }

    this.shouldRefresh = true
  }

  shouldComponentUpdate(newProps) {
    if (newProps.src != this.props.src) this.shouldRefresh = true

    return true
  }

  render() {
    if (this.shouldRefresh) {
      this.shouldRefresh = false

      fetchProtectedImage(this.props.src).then(imgSrc => {
        this.setState({
          imgSrc,
        })
      })
    }

    return <img {...this.props} src={this.state.imgSrc} />
  }
}

ProtectedImage.propTypes = {
  src: PropTypes.string.isRequired,
}

export default ProtectedImage
