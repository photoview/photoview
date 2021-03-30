const lazyLoad = elements => {
  function setSrcAttribute(element) {
    if (element.hasAttribute('data-src')) {
      const src = element.getAttribute('data-src')
      element.removeAttribute('data-src')
      element.setAttribute('src', src)
    }
  }

  function loadImage(observer, element) {
    setSrcAttribute(element)
    observer.unobserve(element)
  }

  const images = Array.from(elements)
  if (images.length) {
    if (!('IntersectionObserver' in window)) {
      images.forEach(image => setSrcAttribute(image))
    } else {
      const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting || entry.intersectionRatio > 0) {
            loadImage(observer, entry.target)
          }
        })
      })
      images.forEach(image => observer.observe(image))
    }
  }
}

export default lazyLoad
