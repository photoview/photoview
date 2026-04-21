import { render, screen } from '@testing-library/react'
import React from 'react'
import Layout from './Layout'
import * as AuthorizedRoute from '../routes/AuthorizedRoute'

describe('Layout component', () => {

  test('with left menu bar', () => {
    vi.spyOn(AuthorizedRoute, 'useIsAuthorized').mockReturnValue(true)

    render(
        <Layout title="Test title">
          <p>layout_content</p>
        </Layout>
    )

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    const p = screen.getByText('layout_content')
    expect(p).toBeInTheDocument()

    const div = p.parentElement
    expect(div).toHaveClass('lg:ml-[292px]')
  })

  test('without left menu bar', () => {
    vi.spyOn(AuthorizedRoute, 'useIsAuthorized').mockReturnValue(false)

    render(
        <Layout title="Test title">
          <p>layout_content</p>
        </Layout>
    )

    expect(screen.getByTestId('Layout')).toBeInTheDocument()
    const p = screen.getByText('layout_content')
    expect(p).toBeInTheDocument()

    const div = p.parentElement
    expect(div).not.toHaveClass('lg:ml-[292px]')
  })
})
