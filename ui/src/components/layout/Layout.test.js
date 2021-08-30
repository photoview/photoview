import { render, screen } from '@testing-library/react'
import React from 'react'
import Layout from './Layout'

test('Layout component', async () => {
  render(
    <Layout>
      <p>layout_content</p>
    </Layout>
  )

  expect(screen.getByTestId('Layout')).toBeInTheDocument()
  expect(screen.getByText('layout_content')).toBeInTheDocument()
})
