import { MockedProvider } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import React from 'react'
import Layout from './Layout'

test('Layout component', () => {
  render(
    <MockedProvider mocks={[]}>
      <Layout title="Test title">
        <p>layout_content</p>
      </Layout>
    </MockedProvider>
  )

  expect(screen.getByTestId('Layout')).toBeInTheDocument()
  expect(screen.getByText('layout_content')).toBeInTheDocument()
})
