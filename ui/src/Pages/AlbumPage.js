import React, { Component } from 'react'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import styled from 'styled-components'
import { useSpring, animated } from 'react-spring'

const albumQuery = gql`
  query albumQuery($id: ID) {
    album(id: $id) {
      title
      photos {
        id
        thumbnail {
          path
          width
          height
        }
      }
    }
  }
`

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

const Photo = styled.img`
  margin: 4px;
  background-color: #eee;
  display: inline-block;
  height: 200px;
  flex-grow: 1;
  object-fit: cover;

  &:nth-last-child(6) ~ img {
    flex-grow: 0;
  }

  ${props =>
    props.active &&
    `
      will-change: transform;
      position: relative;
      z-index: 999;
    `}
`

const Dimmer = ({ onClick, active }) => {
  const [props, set, stop] = useSpring(() => ({ opacity: 0 }))

  set({
    opacity: active ? 1 : 0,
  })

  const AnimatedDimmer = styled(animated.div)`
    position: fixed;
    width: 100%;
    height: 100%;
    background-color: black;
    margin: 0;
    z-index: 10;
  `

  return (
    <AnimatedDimmer
      onClick={onClick}
      style={{
        ...props,
        pointerEvents: active ? 'auto' : 'none',
      }}
    />
  )
}

class AlbumPage extends Component {
  constructor(props) {
    super(props)

    this.state = {
      activeImage: -1,
    }

    this.setActiveImage = this.setActiveImage.bind(this)
    this.animateImage = this.animateImage.bind(this)

    this.photoAmount = 1
    this.previousActive = false

    this.scrollEvent = () => {
      if (this.state.activeImage != -1) {
        this.dismissPopup()
      }
    }

    this.keyUpEvent = e => {
      const activeImage = this.state.activeImage
      if (activeImage != -1) {
        if (e.key == 'ArrowRight') {
          this.setActiveImage((activeImage + 1) % this.photoAmount)
        }

        if (e.key == 'ArrowLeft') {
          if (activeImage <= 0) {
            this.setActiveImage(this.photoAmount - 1)
          } else {
            this.setActiveImage(activeImage - 1)
          }
        }
      }
    }
  }

  componentDidMount() {
    document.addEventListener('scroll', this.scrollEvent)
    document.addEventListener('keyup', this.keyUpEvent)
  }

  componentWillUnmount() {
    document.removeEventListener('scroll', this.scrollEvent)
    document.removeEventListener('keyup', this.keyUpEvent)
  }

  setActiveImage(index) {
    this.previousActive = this.state.activeImage != -1
    this.setState({
      activeImage: index,
    })
  }

  dismissPopup() {
    this.previousActive = true
    this.setState({
      activeImage: -1,
    })
  }

  animateImage(target) {
    const margin = 32

    const windowWidth = window.innerWidth
    const windowHeight = window.innerHeight

    const viewportWidth = windowWidth - margin * 2
    const viewportHeight = windowHeight - margin * 2

    const { naturalWidth, naturalHeight } = target
    const { width, height, top, left } = target.getBoundingClientRect()

    const scaleX = Math.min(naturalWidth, viewportWidth) / width
    const scaleY = Math.min(naturalHeight, viewportHeight) / height
    const scale = Math.min(scaleX, scaleY)

    const translateX = (-left + (windowWidth - width) / 2) / scale
    const translateY = (-top + (windowHeight - height) / 2) / scale

    if (!this.previousActive) target.style.transition = `transform 0.4s`

    requestAnimationFrame(() => {
      target.style.transform = `scale(${scale})
      translate3d(${translateX}px, ${translateY}px, 0)`
    })
  }

  render() {
    const albumId = this.props.match.params.id

    // let dimmer = this.state.activeImage != -1 && (
    //   <Dimmer onClick={() => this.setActiveImage(-1)} />
    // )

    return (
      <div>
        <Dimmer
          onClick={() => this.setActiveImage(-1)}
          active={this.state.activeImage != -1}
        />
        <Query query={albumQuery} variables={{ id: albumId }}>
          {({ loading, error, data }) => {
            if (error) return <div>Error</div>
            if (loading) return <div>Loading</div>

            this.photoAmount = data.album.photos.length

            const { activeImage } = this.state

            const photos = data.album.photos.map((photo, index) => {
              const active = index == activeImage
              let ref = null
              let style = {}

              if (active) {
                ref = target => {
                  if (target == null) return
                  this.animateImage(target)
                }
              } else {
                style.transform = null
              }

              return (
                <Photo
                  ref={ref}
                  active={active}
                  key={photo.id}
                  onClick={() => {
                    requestAnimationFrame(() => {
                      if (activeImage != index) {
                        this.setActiveImage(index)
                      } else {
                        this.dismissPopup()
                      }
                    })
                  }}
                  style={style}
                  src={photo.thumbnail.path}
                ></Photo>
              )
            })

            return <Gallery>{photos}</Gallery>
          }}
        </Query>
      </div>
    )
  }
}

export default AlbumPage
