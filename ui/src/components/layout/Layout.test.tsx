import { render, screen } from '@testing-library/react'
import React from 'react'
import Layout from './Layout'
import * as authentication from "../../helpers/authentication";

vi.mock('../../helpers/authentication.ts')

const isAuthorized = vi.mocked(authentication.isAuthorized)

describe('Layout component', () => {

  test('with left menu bar', () => {
    isAuthorized.mockImplementation(() => true)

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
    isAuthorized.mockImplementation(() => false)

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
