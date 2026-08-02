export const updateTheme = () => {
  if (
    localStorage.theme === 'dark' ||
    (!('theme' in localStorage) &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

export const changeTheme = (value: string) => {
  if (value == 'light') {
    localStorage.theme = 'light'
  } else if (value == 'dark') {
    localStorage.theme = 'dark'
  } else {
    // use OS preference
    localStorage.removeItem('theme')
  }

  updateTheme()
}

export const getTheme = () => {
  if (localStorage.theme == 'light') {
    return 'light'
  } else if (localStorage.theme == 'dark') {
    return 'dark'
  } else {
    return 'auto'
  }
}

export const isDarkMode = () =>
  document.documentElement.classList.contains('dark')

export const SWIPE_NAVIGATION_KEY = 'swipeNavigationDisabled'

export const getSwipeNavigationDisabled = (): boolean => {
  return localStorage.getItem(SWIPE_NAVIGATION_KEY) === 'true'
}

export const setSwipeNavigationDisabled = (disabled: boolean) => {
  if (disabled) {
    localStorage.setItem(SWIPE_NAVIGATION_KEY, 'true')
  } else {
    localStorage.removeItem(SWIPE_NAVIGATION_KEY)
  }
}
