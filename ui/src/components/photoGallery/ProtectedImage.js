import React, { useState, useEffect } from 'react'

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

  async fetchImage() {
    if (this.props.src && this.shouldRefresh) {
      this.shouldRefresh = false

      let image = await fetch(this.props.src, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      })

      image = await image.blob()
      const url = URL.createObjectURL(image)

      this.setState({
        imgSrc: url,
        loadedSrc: this.props.src,
      })
    }
  }

  render() {
    this.fetchImage()
    return <img {...this.props} src={this.state.imgSrc} />
  }
}

export default ProtectedImage
