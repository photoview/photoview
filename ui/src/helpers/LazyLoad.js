class LazyLoad {
  constructor() {
    this.observe = this.observe.bind(this)
    this.loadImages = this.loadImages.bind(this)
    this.disconnect = this.disconnect.bind(this)
    this.observer = null
  }

  observe(images) {
    if (!this.observer) {
      this.observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting || entry.intersectionRatio > 0) {
            const element = entry.target
            this.setSrcAttribute(element)
            this.observer.unobserve(element)
          }
        })
      })
    }
    Array.from(images).forEach(image => this.observer.observe(image))
  }

  loadImages(elements) {
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

  setSrcAttribute(element) {
    if (element.hasAttribute('data-src')) {
      const src = element.getAttribute('data-src')
      element.removeAttribute('data-src')
      element.setAttribute('src', src)
    }
  }
}

export default new LazyLoad()
