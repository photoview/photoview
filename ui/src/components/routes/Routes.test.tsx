import React from 'react'

import Routes from './Routes'
import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

vi.mock('../../Pages/LoginPage/LoginPage.tsx', () => () => (
  <div>mocked login page</div>
))

describe('routes', () => {
  // vitest does not support this yet
  // https://github.com/vitest-dev/vitest/issues/960
  test.skip('unauthorized root path should navigate to login page', async () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes />
      </MemoryRouter>
    )

    await waitForElementToBeRemoved(() =>
      screen.getByText('Loading', { exact: false })
    )

    expect(screen.getByText('mocked login page')).toBeInTheDocument()
  })

  test('invalid page should print a "not found" message', () => {
    render(
      <MemoryRouter initialEntries={['/random_non_existent_page']}>
        <Routes />
      </MemoryRouter>
    )

    expect(screen.getByText('Page not found')).toBeInTheDocument()
  })
})
