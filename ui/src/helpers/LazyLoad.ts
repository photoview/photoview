class LazyLoad {
  observer: null | IntersectionObserver

  constructor() {
    this.observe = this.observe.bind(this)
    this.loadImages = this.loadImages.bind(this)
    this.disconnect = this.disconnect.bind(this)
    this.observer = null
  }

  observe(images: Element[]) {
    if (!this.observer) {
      this.observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting || entry.intersectionRatio > 0) {
            const element = entry.target
            this.setSrcAttribute(element)
            this.observer?.unobserve(element)
          }
        })
      })
    }
    Array.from(images).forEach(image => this.observer?.observe(image))
  }

  loadImages(elements: Element[]) {
    const images = Array.from(elements)
    if (images.length) {
      if ('IntersectionObserver' in window) {
        this.observe(images)
      } else {
        images.forEach(image => this.setSrcAttribute(image))
      }
    }
  }

  disconnect() {
    this.observer && this.observer.disconnect()
  }

  setSrcAttribute(element: Element) {
    if (element.hasAttribute('data-src')) {
      const src = element.getAttribute('data-src')
      if (src) {
        element.removeAttribute('data-src')
        element.setAttribute('src', src)
      } else {
        console.warn(
          'WARN: expected element to have `data-src` property',
          element
        )
      }
    }
  }
}

export default new LazyLoad()
